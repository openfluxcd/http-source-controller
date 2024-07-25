/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/fluxcd/pkg/runtime/patch"
	artifactv1 "github.com/openfluxcd/artifact/api/v1alpha1"
	"github.com/openfluxcd/controller-manager/storage"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	openfluxcdv1alpha1 "github.com/openfluxcd/http-source-controller/api/v1alpha1"
	"github.com/openfluxcd/http-source-controller/internal/fetcher"
)

// HttpReconciler reconciles a Http object
type HttpReconciler struct {
	client.Client
	Scheme *runtime.Scheme

	Fetcher *fetcher.Fetcher
	Storage *storage.Storage
}

//+kubebuilder:rbac:groups=openfluxcd.openfluxcd,resources=https,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=openfluxcd.openfluxcd,resources=https/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=openfluxcd.openfluxcd,resources=https/finalizers,verbs=update
//+kubebuilder:rbac:groups=openfluxcd.mandelsoft.org,resources=artifacts,verbs=get;list;watch;create;update;patch;delete

// Reconcile loop.
func (r *HttpReconciler) Reconcile(ctx context.Context, req ctrl.Request) (_ ctrl.Result, retErr error) {
	logger := log.FromContext(ctx).WithName("http-source-controller")

	logger.Info("starting http source controller loop")

	obj := &openfluxcdv1alpha1.Http{}
	if err := r.Client.Get(ctx, req.NamespacedName, obj); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, fmt.Errorf("failed to get component object: %w", err)
	}

	if obj.GetDeletionTimestamp() != nil {
		logger.Info("deleting http source controller")
		return ctrl.Result{}, nil
	}

	patchHelper := patch.NewSerialPatcher(obj, r.Client)

	// Always attempt to patch the object and status after each reconciliation.
	defer func() {
		if perr := patchHelper.Patch(ctx, obj); perr != nil {
			retErr = errors.Join(retErr, perr)
		}
	}()

	// Create temp working dir
	tmpDir, err := os.MkdirTemp("", fmt.Sprintf("%s-%s-%s-", obj.Kind, obj.Namespace, obj.Name))
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to create temporary working directory: %w", err)
	}
	defer func() {
		if err = os.RemoveAll(tmpDir); err != nil {
			ctrl.LoggerFrom(ctx).Error(err, "failed to remove temporary working directory")
		}
	}()

	// reconcile the source and put it into the folder that the archive is going to serve.
	digest, err := r.Fetcher.Fetch(ctx, obj.Spec.URL, tmpDir)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to fetch http source: %w", err)
	}

	artifact, err := r.findArtifact(ctx, obj)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to find artifact: %w", err)
	}

	if err := r.Storage.ReconcileStorage(ctx, obj, artifact); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to reconcile storage: %w", err)
	}

	// It's important to create this AFTER we reconcile the storage.
	if artifact == nil {
		logger.Info("artifact not found")
		genArt := r.Storage.NewArtifactFor("Http", obj.GetObjectMeta(), digest, digest+".tar.gz")
		artifact = &genArt
	}

	// Revision here is the hash of the content of the downloaded file for example.
	if err := r.Storage.ReconcileArtifact(ctx, obj, artifact, digest, tmpDir, digest+".tar.gz", func(artifact *artifactv1.Artifact, s string) error {
		// Archive directory to storage
		if err := r.Storage.Archive(artifact, tmpDir, nil); err != nil {
			return fmt.Errorf("unable to archive artifact to storage: %w", err)
		}

		return nil
	}); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to reconcile artifact: %w", err)
	}

	if _, err = controllerutil.CreateOrUpdate(ctx, r.Client, artifact, func() error {
		if artifact.ObjectMeta.CreationTimestamp.IsZero() {
			if err := controllerutil.SetOwnerReference(obj, artifact, r.Scheme); err != nil {
				return fmt.Errorf("failed to set owner reference: %w", err)
			}
		}

		artifact.Spec.Revision = digest
		artifact.Spec.Digest = digest

		return nil
	}); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to create/update artifact: %w", err)
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *HttpReconciler) SetupWithManager(_ context.Context, mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&openfluxcdv1alpha1.Http{}, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		Complete(r)
}

// this should most likely be extracted into the controller-manager
func (r *HttpReconciler) findArtifact(ctx context.Context, object client.Object) (*artifactv1.Artifact, error) {
	for _, ch := range object.GetOwnerReferences() {
		if ch.Kind == artifactv1.ArtifactKind {
			artifact := &artifactv1.Artifact{}
			if err := r.Get(ctx, types.NamespacedName{Name: ch.Name, Namespace: object.GetNamespace()}, artifact); err != nil {
				if apierrors.IsNotFound(err) {
					return artifact, nil
				}

				return nil, fmt.Errorf("failed to get artifact: %w", err)
			}

			return artifact, nil
		}
	}

	return nil, nil
}

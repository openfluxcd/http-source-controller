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

	"github.com/fluxcd/pkg/runtime/patch"
	openfluxcdv1alpha1 "github.com/openfluxcd/http-source-controller/api/v1alpha1"
	"github.com/openfluxcd/http-source-controller/internal/fetcher"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// HttpReconciler reconciles a Http object
type HttpReconciler struct {
	client.Client
	Scheme  *runtime.Scheme
	Fetcher *fetcher.Fetcher
}

//+kubebuilder:rbac:groups=openfluxcd.openfluxcd,resources=https,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=openfluxcd.openfluxcd,resources=https/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=openfluxcd.openfluxcd,resources=https/finalizers,verbs=update

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

	content, err := r.Fetcher.Fetch(ctx, obj.Spec.URL)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to fetch http source: %w", err)
	}

	logger.Info("fetched", "content", string(content))

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *HttpReconciler) SetupWithManager(_ context.Context, mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&openfluxcdv1alpha1.Http{}, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		Complete(r)
}

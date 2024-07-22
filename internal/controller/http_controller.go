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

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	openfluxcdv1alpha1 "github.com/openfluxcd/http-controller/api/v1alpha1"
)

// HttpReconciler reconciles a Http object
type HttpReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=openfluxcd.openfluxcd,resources=https,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=openfluxcd.openfluxcd,resources=https/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=openfluxcd.openfluxcd,resources=https/finalizers,verbs=update

// Reconcile loop.
func (r *HttpReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// TODO(user): your logic here

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *HttpReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&openfluxcdv1alpha1.Http{}).
		Complete(r)
}

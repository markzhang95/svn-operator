/*
Copyright 2024 markzhang95.

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
	"reflect"
	"time"

	tappsv1 "github.com/markzhang95/application-operator/api/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

var CounterReconcileApplication int64

const GenericRequeueDuration = 1 * time.Minute

// ApplicationReconciler reconciles a Application object
type ApplicationReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=apps.zhangyi.chat,resources=applications,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps.zhangyi.chat,resources=applications/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=apps.zhangyi.chat,resources=applications/finalizers,verbs=update

// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=services/status,verbs=get

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Application object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.17.0/pkg/reconcile
func (r *ApplicationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	<-time.NewTicker(100 * time.Millisecond).C
	l := log.FromContext(ctx)

	CounterReconcileApplication += 1
	l.Info("Starting a reconcile", "number", CounterReconcileApplication)
	// get the Application
	app := &tappsv1.Application{}
	if err := r.Get(ctx, req.NamespacedName, app); err != nil {
		if errors.IsNotFound(err) {
			l.Info("the Application is not found")
			return ctrl.Result{}, nil
		}
		l.Error(err, "failed to get the Application， will requeue after a short time.")
		return ctrl.Result{RequeueAfter: GenericRequeueDuration}, err
	}
	// reconcile sub-resources
	var result ctrl.Result
	var err error
	result, err = r.ReconcileDeployment(ctx, app)
	if err != nil {
		l.Error(err, "Failed to reconcile Deployment.")
		return result, err
	}

	result, err = r.ReconcileService(ctx, app)
	if err != nil {
		l.Error(err, "Failed to reconcile Service.")
		return result, err
	}
	l.Info("All resources have been reconciled.")
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ApplicationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	setupLog := ctrl.Log.WithName("setup")

	return ctrl.NewControllerManagedBy(mgr).
		For(&tappsv1.Application{}, builder.WithPredicates(predicate.Funcs{
			CreateFunc: func(event event.CreateEvent) bool {
				return true
			},
			DeleteFunc: func(deleteEvent event.DeleteEvent) bool {
				setupLog.Info("The Application has been deleted.",
					"name", deleteEvent.Object.GetName())
				return false
			},
			UpdateFunc: func(updateEvent event.UpdateEvent) bool {
				if updateEvent.ObjectNew.GetResourceVersion() == updateEvent.ObjectOld.GetResourceVersion() {
					return false
				}
				if reflect.DeepEqual(updateEvent.ObjectNew.(*tappsv1.Application).Spec,
					updateEvent.ObjectOld.(*tappsv1.Application).Spec) {
					return false
				}
				return true
			},
		})).
		Owns(&appsv1.Deployment{}, builder.WithPredicates(predicate.Funcs{
			CreateFunc: func(createEvent event.CreateEvent) bool {
				return false
			},
			DeleteFunc: func(deleteEvent event.DeleteEvent) bool {
				setupLog.Info("The Deployment has been deleted.",
					"name", deleteEvent.Object.GetName())
				return true
			},
			UpdateFunc: func(updateEvent event.UpdateEvent) bool {
				if updateEvent.ObjectNew.GetResourceVersion() == updateEvent.ObjectOld.GetResourceVersion() {
					return false
				}
				if reflect.DeepEqual(updateEvent.ObjectNew.(*appsv1.Deployment).Spec,
					updateEvent.ObjectOld.(*appsv1.Deployment).Spec) {
					return false
				}
				return true
			},
			GenericFunc: nil,
		})).
		Owns(&corev1.Service{}, builder.WithPredicates(predicate.Funcs{
			CreateFunc: func(createEvent event.CreateEvent) bool {
				return false
			},
			DeleteFunc: func(deleteEvent event.DeleteEvent) bool {
				setupLog.Info("The Service has been deleted.",
					"name", deleteEvent.Object.GetName())
				return true
			},
			UpdateFunc: func(updateEvent event.UpdateEvent) bool {
				if updateEvent.ObjectNew.GetResourceVersion() == updateEvent.ObjectOld.GetResourceVersion() {
					return false
				}
				if reflect.DeepEqual(updateEvent.ObjectNew.(*tappsv1.Application).Spec,
					updateEvent.ObjectOld.(*tappsv1.Application).Spec) {
					return false
				}
				return true
			},
			GenericFunc: nil,
		})).
		Complete(r)
}

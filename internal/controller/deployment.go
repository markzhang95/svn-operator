package controller

import (
	"context"
	"reflect"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"

	tappsv1 "github.com/markzhang95/application-operator/api/v1"
)

func (r *ApplicationReconciler) ReconcileDeployment(ctx context.Context, app *tappsv1.Application) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	dp := &appsv1.Deployment{}
	err := r.Get(ctx, types.NamespacedName{
		Namespace: app.Namespace,
		Name:      app.Name,
	}, dp)
	if err == nil {
		l.Info("The Deployment has already exist.")
		if reflect.DeepEqual(dp.Status, app.Status.Workflow) {
			return ctrl.Result{}, nil
		}
		app.Status.Workflow = dp.Status
		if err = r.Status().Update(ctx, app); err != nil {
			l.Error(err, "Failed to update Application status")
			return ctrl.Result{RequeueAfter: GenericRequeueDuration}, err
		}
		l.Info("The Application status has been updated.")
		return ctrl.Result{}, nil
	}
	if !errors.IsNotFound(err) {
		l.Error(err, "Failed to get Deployment, will requeue after a short time.")
		return ctrl.Result{RequeueAfter: GenericRequeueDuration}, err
	}
	newDp := &appsv1.Deployment{}
	newDp.SetName(app.Name)
	newDp.SetNamespace(app.Namespace)
	newDp.SetLabels(app.Labels)
	newDp.Spec = app.Spec.Deployment.DeploymentSpec
	newDp.Spec.Template.SetLabels(app.Labels)

	if err = ctrl.SetControllerReference(app, newDp, r.Scheme); err != nil {
		l.Error(err, "Failed to SetControllerReference, will requeue after a short time.")
		return ctrl.Result{RequeueAfter: GenericRequeueDuration}, err
	}
	if err = r.Create(ctx, newDp); err != nil {
		l.Error(err, "Failed to create Deployment, will requeue after a short time.")
		return ctrl.Result{RequeueAfter: GenericRequeueDuration}, err
	}
	l.Info("The Deployment has been created.")
	return ctrl.Result{}, nil
}

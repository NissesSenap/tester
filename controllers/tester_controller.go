/*
Copyright 2021.

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

package controllers

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	testerv1alpha1 "flagger.app/tester/api/v1alpha1"
)

// TesterReconciler reconciles a Tester object
type TesterReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

const (
	containerName = "tester"
	imageName     = "quay.io/tester:latest"
)

//+kubebuilder:rbac:groups=tester.flagger.app,resources=testers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=tester.flagger.app,resources=testers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=tester.flagger.app,resources=testers/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Tester object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *TesterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	var tester testerv1alpha1.Tester
	if err := r.Get(ctx, req.NamespacedName, &tester); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	/*
		Get a list of current Testers and check the status against it.
		Need to be able to compare the specs towards a existing deployment.
		if not found create it, make sure that we set owenerRef correct on the deployment.
			For now lets use a hello world thing with a sleep.
		Continue from there...
	*/
	return ctrl.Result{}, nil
}

func (r *TesterReconciler) GetTesterDeployment(ctx context.Context, req ctrl.Request) (*appsv1.Deployment, error) {

	// TODO (Edvin) understand how to actually get the object, I need to verify against it.
	err := r.Client.Get(ctx, client.ObjectKey{
		Namespace: req.Namespace,
		Name:      req.Name,
	}, &appsv1.Deployment{})
	if err != nil {
		return &appsv1.Deployment{}, err
	}
	return &appsv1.Deployment{}, nil
}

func (r *TesterReconciler) CreateTesterDeployment(ctx context.Context, tester testerv1alpha1.Tester) error {
	// TODO(Edvin): How should I send around ctx?, probably in the struct...

	deploymentName := tester.ObjectMeta.Name + "-deployment"

	deploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      tester.ObjectMeta.Name + "-deployment",
			Namespace: tester.ObjectMeta.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"deployment": deploymentName},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"deployment": deploymentName}},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: containerName,
							// TODO (Edvin) enable the user to define he's own deployment spec
							Image: imageName,
						},
					},
				},
			},
		},
	}
	err := r.Create(ctx, deploy)
	if err != nil {
		return err
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *TesterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&testerv1alpha1.Tester{}).
		Complete(r)
}

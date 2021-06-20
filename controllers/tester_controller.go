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
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	testerv1alpha1 "flagger.app/tester/api/v1alpha1"
)

// TesterReconciler reconciles a Tester object
type TesterReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	//K8sClient kubernetes.Interface
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
	logger := log.FromContext(ctx)

	var tester testerv1alpha1.Tester
	if err := r.Get(ctx, req.NamespacedName, &tester); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// This is the correct Deployment according to the CR, compare it to the real deployed one
	deployment, err := r.ParseConfig(tester)
	if err != nil {
		return ctrl.Result{}, err
	}

	// Check for existing deployment
	_, err = r.GetTesterDeployment(ctx, req)
	//		https://book-v1.book.kubebuilder.io/basics/simple_controller.html

	if err != nil && errors.IsNotFound(err) {
		logger.Info("Creating Deployment %s/%s\n", deployment.Namespace, deployment.Name)
		err = r.Create(ctx, deployment)
		if err != nil {
			return reconcile.Result{}, err
		}
	}

	// TODO(Edvin) How should i compare the data with the existing deployment

	return ctrl.Result{}, nil
}

// ParsesConfig check the config and generates how the deployment should look
func (r *TesterReconciler) ParseConfig(tester testerv1alpha1.Tester) (*appsv1.Deployment, error) {
	// Need some kind of asbtraction to forward what kind of config that we want to check
	// Tekton, Webhook, etc. doing if for each isn't nice.

	deployment := &appsv1.Deployment{}
	var err error
	if tester.Spec.Tekton != nil {
		deployment, err = r.parseTektonConfig(tester)
		if err != nil {
			return &appsv1.Deployment{}, err
		}
	}

	//TODO(Edvin) Add all the generic stuff like certificate, blocking, ownerRef? Or is that done by the controller?
	return deployment, nil
}

func (r *TesterReconciler) parseTektonConfig(tester testerv1alpha1.Tester) (*appsv1.Deployment, error) {
	deploymentName := tester.ObjectMeta.Name + "-deployment"

	// TODO(Edvin) shoulden't define this much, instead we should patch it in for each config change to make the whole thing
	namespace := tester.ObjectMeta.Namespace

	// TODO(Edvin), what will happen if this value is not defined? Have to write test to understand.
	if tester.Spec.Tekton.Namespace != "" {
		namespace = tester.Spec.Tekton.Namespace
	}

	deploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deploymentName,
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
							Args: []string{"--kind=tekton",
								"--URL=" + tester.Spec.Tekton.Url,
								"--namespace=" + namespace,
							},
						},
					},
				},
			},
		},
	}

	return deploy, nil
}

// GetTesterDeployment gets the current deployment connected with the CR request.
func (r *TesterReconciler) GetTesterDeployment(ctx context.Context, req ctrl.Request) (*appsv1.Deployment, error) {
	var deployment2 appsv1.Deployment
	deploymentNamespace := types.NamespacedName{
		Namespace: req.Namespace,
		Name:      req.Name + "-deployment",
	}
	err := r.Get(ctx, deploymentNamespace, &deployment2)
	if err != nil {
		return &appsv1.Deployment{}, err
	}

	/*
		TODO(Edvin) remove this, when it feels okay.
		deployment, err := r.K8sClient.AppsV1().Deployments(req.Namespace).Get(ctx, req.Name, metav1.GetOptions{})
		if err != nil {
			return &appsv1.Deployment{}, err
		}
	*/
	return &deployment2, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *TesterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&testerv1alpha1.Tester{}).
		Complete(r)
}

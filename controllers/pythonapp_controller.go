/*
Copyright 2022.

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

	"sigs.k8s.io/controller-runtime/pkg/log"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8serros "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	testv1alpha1 "github.com/vishwajitdandage/go-operator/api/v1alpha1"
)

var _log = logf.Log.WithName("controller_traveller")

// PythonAppReconciler reconciles a PythonApp object
type PythonAppReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=test.example.com,resources=pythonapps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=test.example.com,resources=pythonapps/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=test.example.com,resources=pythonapps/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the PythonApp object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *PythonAppReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// TODO(user): your logic here
	//
	appInstance := &testv1alpha1.PythonApp{}
	err := r.Client.Get(ctx, req.NamespacedName, appInstance)
	if err != nil {
		return ctrl.Result{}, err
	}
	var result *ctrl.Result

	result, err = r.reconcileDeployment(req, appInstance, r.createDeployment(appInstance))
	if result != nil {
		return *result, err
	}

	result, err = r.reconcileService(req, appInstance, r.createService(appInstance))
	if result != nil {
		return *result, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PythonAppReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&testv1alpha1.PythonApp{}).
		Complete(r)
}

// Deployment
func (r *PythonAppReconciler) createDeployment(app *testv1alpha1.PythonApp) *appsv1.Deployment {
	replicas := app.Spec.Replicas
	image := app.Spec.Image

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "python-app",
			Namespace: app.Namespace,
			Labels: map[string]string{
				"app": "flaskcalculator",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "flaskcalculator",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "flaskcalculator",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Image: image,
							Name:  "flask-demo",
							Ports: []corev1.ContainerPort{
								{ContainerPort: 5000},
							},
						},
					},
				},
			},
		},
	}
	controllerutil.SetControllerReference(app, dep, r.Scheme)
	return dep
}

// Reconcile Deployment
func (r *PythonAppReconciler) reconcileDeployment(req reconcile.Request, app *testv1alpha1.PythonApp, dep *appsv1.Deployment) (*ctrl.Result, error) {

	found := &appsv1.Deployment{}
	err := r.Client.Get(context.TODO(), types.NamespacedName{Name: dep.Name, Namespace: app.Namespace}, found)
	if err != nil && k8serros.IsNotFound(err) {
		_log.Info("Creating a new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
		err = r.Client.Create(context.TODO(), dep)
		if err != nil {
			// Deployment failed
			_log.Error(err, "Failed to create new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
			return &reconcile.Result{}, err
		} else {
			// Deployment was successful
			return nil, nil
		}
	} else if err != nil {
		// Error that isn't due to the deployment not existing
		_log.Error(err, "Failed to get Deployment")
		return &reconcile.Result{}, err
	}
	return nil, nil
}

func (r *PythonAppReconciler) createService(app *testv1alpha1.PythonApp) *corev1.Service {

	labels := map[string]string{
		"app": "flaskcalculator",
	}
	s := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "flaskservice",
			Namespace: app.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{{
				Protocol:   corev1.ProtocolTCP,
				Port:       5000,
				TargetPort: intstr.FromInt(5000),
			}},
			Type: corev1.ServiceTypeNodePort,
		},
	}
	return s
}

// Reconcile Service
func (r *PythonAppReconciler) reconcileService(req reconcile.Request, app *testv1alpha1.PythonApp, svc *corev1.Service) (*ctrl.Result, error) {

	found := &corev1.Service{}
	err := r.Client.Get(context.TODO(), types.NamespacedName{Name: svc.Name, Namespace: app.Namespace}, found)
	if err != nil && k8serros.IsNotFound(err) {
		_log.Info("Creating a new Service", "Service.Namespace", svc.Namespace, "Service.Name", svc.Name)
		err = r.Client.Create(context.TODO(), svc)
		if err != nil {
			// Service failed
			_log.Error(err, "Failed to create new Service", "Service.Namespace", svc.Namespace, "Service.Name", svc.Name)
			return &reconcile.Result{}, err
		} else {
			// Service was successful
			return nil, nil
		}
	} else if err != nil {
		// Error that isn't due to the Service not existing
		_log.Error(err, "Failed to get Service")
		return &reconcile.Result{}, err
	}
	return nil, nil
}

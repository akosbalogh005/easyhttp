/*
Copyright 2023.

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
	"fmt"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	httpapiv1 "github.com/akosbalogh005/easyhttp-operator/api/v1"
)

// for mocking purposes
//
//go:generate mockery --name=ReconcilerClientIF
type ReconcilerClientIF interface {
	client.Reader
	client.Writer
	client.StatusClient
	client.SubResourceClientConstructor
	// Scheme returns the scheme this client is using.
	Scheme() *runtime.Scheme
	// RESTMapper returns the rest this client is using.
	RESTMapper() meta.RESTMapper
}

// for mocking purposes
//
//go:generate mockery --name=SubResourceWriterIF
type SubResourceWriterIF interface {
	client.SubResourceWriter
}

// EasyHttpReconciler reconciles a EasyHttp object
type EasyHttpReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=httpapi.github.com,resources=easyhttps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=httpapi.github.com,resources=easyhttps/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=httpapi.github.com,resources=easyhttps/finalizers,verbs=update
//+kubebuilder:rbac:groups="apps",resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="extensions",resources=ingresses,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="networking.k8s.io",resources=ingresses,verbs=get;list;watch;create;update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the EasyHttp object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *EasyHttpReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// get the resource
	var clientResource = &httpapiv1.EasyHttp{}
	if err := r.Get(ctx, req.NamespacedName, clientResource); err != nil {
		log.Info(fmt.Sprintf("Reconcile loop is running, client may be deleted... Client:%v.%v, Owner:%v, Spec:%v, Status:%v", clientResource.Namespace, clientResource.Name,
			clientResource.OwnerReferences, clientResource.Status, clientResource.Spec))
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	specHasChanged := false
	// spec has changed
	if clientResource.Status.IsDeployOK {
		if !clientResource.Spec.IsEqual(&clientResource.Status.Spec) {
			log.Info(fmt.Sprintf("Spec is differ, reconfigure. Orig: %v, New: %v", clientResource.Spec, clientResource.Status.Spec))
			clientResource.Status.IsDeployOK = false
			clientResource.Status.IsSvcOK = false
			clientResource.Status.IsIngressOK = false
			specHasChanged = true
		}
	}
	clientResource.Spec.DeepCopyInto(&clientResource.Status.Spec)
	err := r.Status().Update(context.TODO(), clientResource)
	if err != nil {
		log.Error(err, "failed to update client status")
		return ctrl.Result{}, err
	}

	log.Info(fmt.Sprintf("Reconcile loop is running... Client:%v.%v, Owner:%v, Spec:%v,  Status:%v", clientResource.Namespace, clientResource.Name,
		clientResource.OwnerReferences, clientResource.Status, clientResource.Spec))

	// 1st step is check if the deployment is ready.
	ret, err := r.CheckDeployment(ctx, req, specHasChanged, clientResource)
	if err != nil {
		return ret, err
	}

	// 2nd step is the service
	ret, svc, err := r.CheckService(ctx, req, specHasChanged, clientResource)
	if err != nil {
		return ret, err
	}

	// 3rd step is the ingress
	ret, err = r.CheckIngress(ctx, req, specHasChanged, clientResource, svc)
	if err != nil {
		return ret, err
	}

	if clientResource.Spec.CertManInssuer == "" {
		log.Info("Certificate manager is disabled. Add certManIssuer to kind spec if necessary")
	} else {
		log.Info(fmt.Sprintf("Using Certificate manager: %v", clientResource.Spec.CertManInssuer))
	}

	return ctrl.Result{}, nil
	//return ctrl.Result{Requeue: true}, nil
	//return ctrl.Result{RequeueAfter: time.Minute}, nil
}

func (r *EasyHttpReconciler) CheckIngress(ctx context.Context, req ctrl.Request, specHasChanged bool, clientResource *httpapiv1.EasyHttp, svc *v1.Service) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	ing := initIngress(clientResource, svc.Name)

	// try to get the current service  ...
	err := r.Client.Get(ctx, client.ObjectKey{Namespace: ing.Namespace, Name: ing.ObjectMeta.Name}, ing)

	isNew := false
	if err != nil && errors.IsNotFound(err) {
		// (re)deploy
		isNew = true
		clientResource.Status.IsIngressOK = false
	} else if err != nil {
		return ctrl.Result{Requeue: true}, fmt.Errorf("cannot get ingress, retying later. %v", err)
	} else {
		// when current found, update the Spec in order to refresh specification if needed
		if specHasChanged {
			newIng := initIngress(clientResource, svc.Name)
			ing.Spec = *newIng.Spec.DeepCopy()
		}
	}

	if !clientResource.Status.IsIngressOK {
		err = r.createOrUpdate(ctx, req, ing, clientResource, &clientResource.Status.IsIngressOK, isNew)
		if err != nil {
			return ctrl.Result{Requeue: true}, fmt.Errorf("failed to create ingress. %v", err)
		}
		log.Info("Ingress has been successfuly created/updated :)")
	}
	log.Info(fmt.Sprintf("Current Ingress is: %v (%v)", ing.Name, ing.UID))
	return ctrl.Result{}, nil
}

func (r *EasyHttpReconciler) CheckService(ctx context.Context, req ctrl.Request, specHasChanged bool, clientResource *httpapiv1.EasyHttp) (ctrl.Result, *v1.Service, error) {
	log := log.FromContext(ctx)
	svc := initService(clientResource)

	// try to get the current service  ...
	err := r.Client.Get(ctx, client.ObjectKey{Namespace: svc.Namespace, Name: svc.ObjectMeta.Name}, svc)

	isNew := false
	if err != nil && errors.IsNotFound(err) {
		// (re)deploy
		isNew = true
		clientResource.Status.IsSvcOK = false
	} else if err != nil {
		return ctrl.Result{Requeue: true}, svc, fmt.Errorf("cannot get service, retying later. %v", err)
	} else {
		// when current found, update the Spec in order to refresh specification if needed
		if specHasChanged {
			newSvc := initService(clientResource)
			svc.Spec = *newSvc.Spec.DeepCopy()
		}
	}

	if !clientResource.Status.IsSvcOK {
		err = r.createOrUpdate(ctx, req, svc, clientResource, &clientResource.Status.IsSvcOK, isNew)
		if err != nil {
			return ctrl.Result{Requeue: true}, svc, fmt.Errorf("failed to create service. %v", err)
		}
		log.Info("Service has been successfuly created/updated :)")
	}
	log.Info(fmt.Sprintf("Current Service is: %v (%v)", svc.Name, svc.UID))
	return ctrl.Result{}, svc, nil
}

func (r *EasyHttpReconciler) CheckDeployment(ctx context.Context, req ctrl.Request, specHasChanged bool, clientResource *httpapiv1.EasyHttp) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// init deployment struct
	dep := initDeployment(clientResource)

	// try to get the current running deployment ...
	err := r.Client.Get(ctx, client.ObjectKey{Namespace: dep.Namespace, Name: dep.ObjectMeta.Name}, dep)

	isNew := false
	if err != nil && errors.IsNotFound(err) {
		// (re)deploy
		isNew = true
		clientResource.Status.IsDeployOK = false
	} else if err != nil {
		return ctrl.Result{Requeue: true}, fmt.Errorf("cannot get deployment, retying later. %v", err)
	} else {
		// when current found, update the Spec in order to frefresh seecification
		if specHasChanged {
			newDep := initDeployment(clientResource)
			dep.Spec = *newDep.Spec.DeepCopy()
		}
	}

	if !clientResource.Status.IsDeployOK {
		err = r.createOrUpdate(ctx, req, dep, clientResource, &clientResource.Status.IsDeployOK, isNew)
		if err != nil {
			return ctrl.Result{Requeue: true}, fmt.Errorf("failed to create deployment. %v", err)
		}
		log.Info("Deployment has been successfuly created/updated :)")
	}
	log.Info(fmt.Sprintf("Current Deployment is: %v (%v)", dep.Name, dep.UID))

	return ctrl.Result{}, nil
}

func (r *EasyHttpReconciler) createOrUpdate(ctx context.Context, req ctrl.Request, obj client.Object, clientResource *httpapiv1.EasyHttp, statusFlag *bool, isNew bool) error {

	log := log.FromContext(ctx)

	// set op as contoller
	err := ctrl.SetControllerReference(clientResource, obj, r.Scheme)
	if err != nil {
		return err
	}

	// let's try to create / update
	if isNew {
		log.Info(fmt.Sprintf("Create new object: %v", obj.GetName()))
		err = r.Create(context.TODO(), obj)
	} else {
		log.Info(fmt.Sprintf("Update object: %v", obj.GetName()))
		err = r.Update(context.TODO(), obj)
	}
	if err != nil {
		return fmt.Errorf("object (%s) has not been created. %v", obj.GetName(), err)
	}

	*statusFlag = true
	err = r.Status().Update(context.TODO(), clientResource)

	if err != nil {
		return fmt.Errorf("failed to update client status. %v", err)
	}
	log.Info(fmt.Sprintf("New Status (object:%s): %v", obj.GetName(), clientResource.Status))

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *EasyHttpReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&httpapiv1.EasyHttp{}).
		//Owns(&appsv1.Deployment{}).
		Complete(r)
}

package controllers

import (
	"context"
	"testing"

	httpapiv1 "github.com/akosbalogh005/easyhttp-operator/api/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/akosbalogh005/easyhttp-operator/controllers/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
)

var clientMock = mocks.ReconcilerClientIF{}
var subResourceWriterMock = mocks.SubResourceWriterIF{}

func setup(t *testing.T) (*EasyHttpReconciler, *ctrl.Request) {
	reconciler := &EasyHttpReconciler{
		Client: &clientMock,
		Scheme: newTestScheme(),
	}
	req := ctrl.Request{}
	req.Namespace = "namespace1"
	req.Name = "kind1"

	return reconciler, &req
}

func newTestScheme() *runtime.Scheme {
	testScheme := runtime.NewScheme()
	_ = httpapiv1.AddToScheme(testScheme)
	_ = appsv1.AddToScheme(testScheme)
	return testScheme
}

// TestDeploymentNew positive test for creating new deployment. No reconfigure. Totally new
func TestDeploymentNewOK(t *testing.T) {

	reconciler, req := setup(t)
	ctx := context.Background()

	clientResource := httpapiv1.EasyHttp{
		//TypeMeta: metav1.TypeMeta{Kind: "EasyHttp"},
		Status: httpapiv1.EasyHttpStatus{
			IsDeployOK: false,
		},
	}

	subResourceWriterMock.On("Update", mock.Anything, mock.Anything).Return(nil).Once()
	clientMock.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(errors.NewNotFound(schema.GroupResource{}, "")).Once()

	newDep := initDeployment(&clientResource)
	err := ctrl.SetControllerReference(&clientResource, newDep, reconciler.Scheme)
	assert.NoError(t, err)

	clientMock.On("Create", mock.Anything, newDep).Return(nil).Once()
	clientMock.On("Status", mock.Anything, mock.Anything).Return(&subResourceWriterMock).Once()
	defer clientMock.AssertExpectations(t)

	res, err := reconciler.CheckDeployment(ctx, *req, false, &clientResource)

	assert.NoError(t, err)
	assert.Equal(t, ctrl.Result{}, res)
}

// TestDeploymentUpdateOK positive test for update deployment. Reconfigure
func TestDeploymentUpdateOK(t *testing.T) {
	reconciler, req := setup(t)
	ctx := context.Background()

	clientResource := httpapiv1.EasyHttp{
		//TypeMeta: metav1.TypeMeta{Kind: "EasyHttp"},
		Status: httpapiv1.EasyHttpStatus{
			IsDeployOK: false,
		},
	}

	subResourceWriterMock.On("Update", mock.Anything, mock.Anything).Return(nil).Once()

	// Get: found
	clientMock.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

	newDep := initDeployment(&clientResource)
	err := ctrl.SetControllerReference(&clientResource, newDep, reconciler.Scheme)
	assert.NoError(t, err)

	clientMock.On("Update", mock.Anything, newDep).Return(nil).Once()
	clientMock.On("Status", mock.Anything, mock.Anything).Return(&subResourceWriterMock).Once()
	defer clientMock.AssertExpectations(t)

	_, err = reconciler.CheckDeployment(ctx, *req, false, &clientResource)

	assert.NoError(t, err)
}

// TestDeploymentAlreadyDeployedOK positive test for already deployed and nothing changed
func TestDeploymentAlreadyDeployedOK(t *testing.T) {
	reconciler, req := setup(t)
	ctx := context.Background()

	clientResource := httpapiv1.EasyHttp{
		//TypeMeta: metav1.TypeMeta{Kind: "EasyHttp"},
		Status: httpapiv1.EasyHttpStatus{
			IsDeployOK: true,
		},
	}

	// Get: found
	clientMock.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
	defer clientMock.AssertExpectations(t)

	_, err := reconciler.CheckDeployment(ctx, *req, false, &clientResource)

	assert.NoError(t, err)
}

// TestServiveNewOK positive test for creating new service. No reconfigure. Totally new
func TestServiveNewOK(t *testing.T) {

	reconciler, req := setup(t)
	ctx := context.Background()

	clientResource := httpapiv1.EasyHttp{
		//TypeMeta: metav1.TypeMeta{Kind: "EasyHttp"},
		Status: httpapiv1.EasyHttpStatus{
			IsSvcOK: false,
		},
	}

	subResourceWriterMock.On("Update", mock.Anything, mock.Anything).Return(nil).Once()
	clientMock.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(errors.NewNotFound(schema.GroupResource{}, "")).Once()
	defer clientMock.AssertExpectations(t)
	defer subResourceWriterMock.AssertExpectations(t)

	newServ := initService(&clientResource)
	err := ctrl.SetControllerReference(&clientResource, newServ, reconciler.Scheme)
	assert.NoError(t, err)

	clientMock.On("Create", mock.Anything, newServ).Return(nil).Once()
	clientMock.On("Status", mock.Anything, mock.Anything).Return(&subResourceWriterMock).Once()

	res, serv, err := reconciler.CheckService(ctx, *req, false, &clientResource)

	assert.NoError(t, err)
	assert.Equal(t, newServ, serv)
	assert.Equal(t, ctrl.Result{}, res)

}

// TestServiceUpdateOK positive test for update service. Reconfigure
func TestServiceUpdateOK(t *testing.T) {
	reconciler, req := setup(t)
	ctx := context.Background()

	clientResource := httpapiv1.EasyHttp{
		//TypeMeta: metav1.TypeMeta{Kind: "EasyHttp"},
		Status: httpapiv1.EasyHttpStatus{
			IsSvcOK: false,
		},
	}

	subResourceWriterMock.On("Update", mock.Anything, mock.Anything).Return(nil).Once()

	// Get: found
	clientMock.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
	defer clientMock.AssertExpectations(t)

	newServ := initService(&clientResource)
	err := ctrl.SetControllerReference(&clientResource, newServ, reconciler.Scheme)
	assert.NoError(t, err)

	clientMock.On("Update", mock.Anything, newServ).Return(nil).Once()
	clientMock.On("Status", mock.Anything, mock.Anything).Return(&subResourceWriterMock).Once()

	res, serv, err := reconciler.CheckService(ctx, *req, false, &clientResource)

	assert.NoError(t, err)
	assert.Equal(t, newServ, serv)
	assert.Equal(t, ctrl.Result{}, res)
}

// TestServiceAlreadyDeployedOK positive test for already deployed and nothing changed
func TestServiceAlreadyDeployedOK(t *testing.T) {
	reconciler, req := setup(t)
	ctx := context.Background()

	clientResource := httpapiv1.EasyHttp{
		//TypeMeta: metav1.TypeMeta{Kind: "EasyHttp"},
		Status: httpapiv1.EasyHttpStatus{
			IsSvcOK: true,
		},
	}

	newServ := initService(&clientResource)
	// Get: found
	clientMock.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
	defer clientMock.AssertExpectations(t)

	res, serv, err := reconciler.CheckService(ctx, *req, false, &clientResource)

	assert.NoError(t, err)
	assert.Equal(t, newServ, serv)
	assert.Equal(t, ctrl.Result{}, res)
}

// TestIngressNewOK positive test for creating new ingress rule. No reconfigure. Totally new
func TestIngressNewOK(t *testing.T) {

	reconciler, req := setup(t)
	ctx := context.Background()

	clientResource := httpapiv1.EasyHttp{
		//TypeMeta: metav1.TypeMeta{Kind: "EasyHttp"},
		Status: httpapiv1.EasyHttpStatus{
			IsIngressOK: false,
		},
	}

	subResourceWriterMock.On("Update", mock.Anything, mock.Anything).Return(nil).Once()
	clientMock.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(errors.NewNotFound(schema.GroupResource{}, "")).Once()
	defer subResourceWriterMock.AssertExpectations(t)
	defer clientMock.AssertExpectations(t)

	newServ := initService(&clientResource)
	newIng := initIngress(&clientResource, newServ.Name)
	err := ctrl.SetControllerReference(&clientResource, newIng, reconciler.Scheme)
	assert.NoError(t, err)

	clientMock.On("Create", mock.Anything, newIng).Return(nil).Once()
	clientMock.On("Status", mock.Anything, mock.Anything).Return(&subResourceWriterMock).Once()

	res, err := reconciler.CheckIngress(ctx, *req, false, &clientResource, newServ)

	assert.NoError(t, err)
	assert.Equal(t, ctrl.Result{}, res)

}

// TestIngressUpdateOK positive test for update ingress. Reconfigure
func TestIngressUpdateOK(t *testing.T) {
	reconciler, req := setup(t)
	ctx := context.Background()

	clientResource := httpapiv1.EasyHttp{
		//TypeMeta: metav1.TypeMeta{Kind: "EasyHttp"},
		Status: httpapiv1.EasyHttpStatus{
			IsIngressOK: false,
		},
	}

	subResourceWriterMock.On("Update", mock.Anything, mock.Anything).Return(nil).Once()
	defer subResourceWriterMock.AssertExpectations(t)
	// Get: found
	clientMock.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
	defer clientMock.AssertExpectations(t)

	newServ := initService(&clientResource)
	newIng := initIngress(&clientResource, newServ.Name)
	err := ctrl.SetControllerReference(&clientResource, newIng, reconciler.Scheme)
	assert.NoError(t, err)

	clientMock.On("Update", mock.Anything, newIng).Return(nil).Once()
	clientMock.On("Status", mock.Anything, mock.Anything).Return(&subResourceWriterMock).Once()

	res, err := reconciler.CheckIngress(ctx, *req, false, &clientResource, newServ)

	assert.NoError(t, err)
	assert.Equal(t, ctrl.Result{}, res)
}

// TestIngressAlreadyDeployedOK positive test for already existed ingress. Nothing changed
func TestIngressAlreadyDeployedOK(t *testing.T) {
	reconciler, req := setup(t)
	ctx := context.Background()

	clientResource := httpapiv1.EasyHttp{
		//TypeMeta: metav1.TypeMeta{Kind: "EasyHttp"},
		Status: httpapiv1.EasyHttpStatus{
			IsIngressOK: true,
		},
	}

	newServ := initService(&clientResource)
	clientMock.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
	defer clientMock.AssertExpectations(t)

	res, err := reconciler.CheckIngress(ctx, *req, false, &clientResource, newServ)

	assert.NoError(t, err)
	assert.Equal(t, ctrl.Result{}, res)
}

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

	_, err = reconciler.CheckDeployment(ctx, *req, false, &clientResource)

	assert.NoError(t, err)
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

	_, err := reconciler.CheckDeployment(ctx, *req, false, &clientResource)

	assert.NoError(t, err)
}

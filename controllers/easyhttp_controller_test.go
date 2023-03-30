package controllers

import (
	"context"
	"testing"

	httpapiv1 "github.com/akosbalogh005/easyhttp-operator/api/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/akosbalogh005/easyhttp-operator/controllers/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	controllerruntime "sigs.k8s.io/controller-runtime"
)

func TestDeploymentNew(t *testing.T) {
	t.Skip()

	clientMock := mocks.ReconcilerClientIF{}
	reconciler := &EasyHttpReconciler{
		Client: &clientMock,
		Scheme: newTestScheme(),
	}
	ctx := context.Background()
	req := controllerruntime.Request{}
	req.Namespace = "namespace1"
	req.Name = "kind1"

	clientResource := httpapiv1.EasyHttp{
		TypeMeta: metav1.TypeMeta{Kind: "EasyHttp"},
		Status: httpapiv1.EasyHttpStatus{
			IsDeployOK: false,
		},
	}

	clientMock.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(errors.NewNotFound(schema.GroupResource{}, "")).Once()
	clientMock.On("Create", mock.Anything, mock.Anything).Return(nil).Once()

	_, err := reconciler.CheckDeployment(ctx, req, true, &clientResource)

	assert.NoError(t, err)
}

func newTestScheme() *runtime.Scheme {
	testScheme := runtime.NewScheme()
	_ = appsv1.AddToScheme(testScheme)
	return testScheme
}

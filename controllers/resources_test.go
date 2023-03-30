package controllers

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"text/template"

	httpapiv1 "github.com/akosbalogh005/easyhttp-operator/api/v1"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func getTempleate(t *testing.T, fileName string) *template.Template {
	depTmp, err := os.ReadFile(filepath.Join("templates", fileName))
	assert.NoError(t, err)
	depTemplate, err := template.New("dep").Parse(string(depTmp))
	assert.NoError(t, err)
	return depTemplate
}

func TestInitDeployment(t *testing.T) {

	clientResource := httpapiv1.EasyHttp{}
	clientResource.Name = "app1"
	clientResource.Namespace = "namespace1"

	var replicas3 int32 = 3

	tests := map[string]httpapiv1.EasyHttpSpec{
		"with env, 3 replicas": {
			Host:           "testhost",
			Replicas:       &replicas3,
			Image:          "testimage",
			ImageTag:       "1.0",
			Port:           1234,
			Env:            map[string]string{"PORT": "1234"},
			CertManInssuer: "local.issuer",
			Path:           "/app",
		},
		"no envs, no replicas": {
			Host:           "testhost2",
			Replicas:       nil,
			Image:          "testimage2",
			ImageTag:       "1.0",
			Port:           1234,
			Env:            nil,
			CertManInssuer: "local.issuer",
			Path:           "/app",
		},
	}

	for k, v := range tests {
		t.Run(k, func(t *testing.T) {

			clientResource.Spec = *v.DeepCopy()
			dep := initDeployment(&clientResource)
			createdYaml, err := yaml.Marshal(dep)
			assert.NoError(t, err)

			var templateYaml bytes.Buffer
			err = getTempleate(t, "deployment.yaml").Execute(&templateYaml, clientResource)
			assert.NoError(t, err)

			assert.Equal(t, templateYaml.String(), string(createdYaml))

		})
	}
}

func TestInitService(t *testing.T) {

	clientResource := httpapiv1.EasyHttp{}
	clientResource.Name = "app1"
	clientResource.Namespace = "namespace1"

	tests := map[string]httpapiv1.EasyHttpSpec{
		"port_1234": {
			Port: 1234,
		},
	}

	for k, v := range tests {
		t.Run(k, func(t *testing.T) {

			clientResource.Spec = *v.DeepCopy()
			dep := initService(&clientResource)
			createdYaml, err := yaml.Marshal(dep)
			assert.NoError(t, err)

			var templateYaml bytes.Buffer
			err = getTempleate(t, "service.yaml").Execute(&templateYaml, clientResource)
			assert.NoError(t, err)
			assert.Equal(t, templateYaml.String(), string(createdYaml))

		})
	}
}

func TestInitIngress(t *testing.T) {

	clientResource := httpapiv1.EasyHttp{}
	clientResource.Name = "app1"
	clientResource.Namespace = "namespace1"

	tests := map[string]httpapiv1.EasyHttpSpec{
		"with path, no issuer": {
			Host:           "testhost",
			Replicas:       nil,
			Image:          "testimage",
			ImageTag:       "1.0",
			Port:           1234,
			Env:            map[string]string{"PORT": "1111", "PORT2": "1234"},
			CertManInssuer: "local.issuer",
			Path:           "/app",
		},
		"no path, with issuer": {
			Host:           "testhost",
			Replicas:       nil,
			Image:          "testimage",
			ImageTag:       "1.0",
			Port:           1234,
			Env:            map[string]string{"PORT": "1111", "PORT2": "1234"},
			CertManInssuer: "local.issuer",
		},
		"no path, no issuer": {
			Host:     "testhost",
			Replicas: nil,
			Image:    "testimage",
			ImageTag: "1.0",
			Port:     1234,
			Env:      map[string]string{"PORT": "1111", "PORT2": "1234"},
		},
	}

	for k, v := range tests {
		t.Run(k, func(t *testing.T) {

			clientResource.Spec = *v.DeepCopy()
			dep := initIngress(&clientResource, clientResource.Name+"-svc")
			createdYaml, err := yaml.Marshal(dep)
			assert.NoError(t, err)

			var templateYaml bytes.Buffer
			err = getTempleate(t, "ingress.yaml").Execute(&templateYaml, clientResource)
			assert.NoError(t, err)

			assert.Equal(t, templateYaml.String(), string(createdYaml))

		})
	}
}

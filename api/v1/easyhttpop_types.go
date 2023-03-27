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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// EasyHttpSpec defines the desired state of EasyHttp
type EasyHttpSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Host where the application is accesible from outside. Base of the Ingress route and certificate request
	Host string `json:"host,omitempty"`
	// Replicas of the HTTP server application
	// +kubebuilder:validation:optional
	Replicas *int32 `json:"replicas,omitempty"`
	// Image of the application
	Image string `json:"image,omitempty"`
	// ImageTag version tag of image
	ImageTag string `json:"tag,omitempty"`
	// Port where the application (and srv) is listening
	Port int `json:"port,omitempty"`
	// Env is the map of environment of the application to be set
	Env map[string]string `json:"env,omitempty"`
	// CertManInssuer issuer of cert manager (e.g 'letsencrypt-prod'). Cert manager is disabled when empty.
	CertManInssuer string `json:"certManIssuer,omitempty"`
	// Path is  where the application can be called (from outside). Currently supported only in nginx ingress!
	// +kubebuilder:validation:optional
	Path string `json:"path,omitempty"`
}

func nvl(v *int32) int32 {
	if v == nil {
		return 1
	}
	return *v
}

func (e *EasyHttpSpec) IsEqual(o *EasyHttpSpec) bool {
	ret := e.Host == o.Host &&
		e.Image == o.Image &&
		e.ImageTag == o.ImageTag &&
		e.Port == o.Port &&
		e.CertManInssuer == o.CertManInssuer &&
		e.Path == o.Path &&
		nvl(e.Replicas) == nvl(o.Replicas)
	if !ret {
		return ret
	}
	if len(e.Env) != len(o.Env) {
		return false
	}
	for k, v := range e.Env {
		v2, has := o.Env[k]
		if !has {
			return false
		}
		if v != v2 {
			return false
		}
	}
	return true

}

// EasyHttpStatus defines the observed state of EasyHttp
type EasyHttpStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// IsDeployOK flag for status of deployment
	IsDeployOK bool `json:"is_deploy_ok,omitempty"`
	// IsSvcOK flag for status of service
	IsSvcOK bool `json:"is_svc_ok,omitempty"`
	// IsIngressOK flag for status of ingress setup
	IsIngressOK bool `json:"is_ingress_ok,omitempty"`
	// IsCertOK flag for status of cert-manager setup
	Spec EasyHttpSpec `json:"spec,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// EasyHttp is the Schema for the easyhttps API
type EasyHttp struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   EasyHttpSpec   `json:"spec,omitempty"`
	Status EasyHttpStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// EasyHttpList contains a list of EasyHttp
type EasyHttpList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []EasyHttp `json:"items"`
}

func init() {
	SchemeBuilder.Register(&EasyHttp{}, &EasyHttpList{})
}

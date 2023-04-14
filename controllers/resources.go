package controllers

import (
	"fmt"
	"strings"

	httpapiv1 "github.com/akosbalogh005/easyhttp-operator/api/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// initService creates service based on clientResource
func initService(clientResource *httpapiv1.EasyHttp) *corev1.Service {
	svc := corev1.Service{}
	var ports []corev1.ServicePort
	ports = append(ports, corev1.ServicePort{Name: "http", Protocol: "TCP", Port: int32(clientResource.Spec.Port)})
	svc.APIVersion = "apps/v1"
	svc.Name = clientResource.Name + "-svc"
	svc.Namespace = clientResource.Namespace
	svc.Spec = corev1.ServiceSpec{Ports: ports}
	svc.Spec.Selector = map[string]string{"app": clientResource.Name}
	return &svc
}

// initDeployment creates deployment based on clientResource
func initDeployment(clientResource *httpapiv1.EasyHttp) *appsv1.Deployment {
	name := clientResource.Name
	var replicas int32 = 1
	if clientResource.Spec.Replicas != nil {
		replicas = *clientResource.Spec.Replicas
	}

	d := appsv1.Deployment{}
	d.APIVersion = "apps/v1"
	d.Kind = "Deployment"
	d.Name = clientResource.Name
	d.Namespace = clientResource.Namespace

	cont := corev1.Container{
		Image: fmt.Sprintf("%s:%s", clientResource.Spec.Image, clientResource.Spec.ImageTag),
		Name:  name,
		Env:   convertEnv(clientResource.Spec.Env),
	}
	cont.Ports = append(cont.Ports, corev1.ContainerPort{
		Name:          name,
		ContainerPort: int32(clientResource.Spec.Port),
	})

	temp := corev1.PodTemplateSpec{}
	temp.Labels = map[string]string{"app": name}
	temp.Spec = corev1.PodSpec{}
	temp.Spec.Containers = append(temp.Spec.Containers, cont)

	d.Spec = appsv1.DeploymentSpec{
		Replicas: &replicas,
		Selector: &metav1.LabelSelector{
			MatchLabels: map[string]string{"app": name},
		},
		Strategy: appsv1.DeploymentStrategy{Type: "Recreate"},
		Template: temp,
	}
	return &d
}

// initIngress creates deployment based on clientResource
func initIngress(clientResource *httpapiv1.EasyHttp, serviceName string) *netv1.Ingress {

	ing := netv1.Ingress{}
	//ing.APIVersion = "networking.k8s.io/v1"
	ing.Kind = "Ingress"
	ing.Name = clientResource.Name + "-ingress"
	ing.Namespace = clientResource.Namespace
	if clientResource.Spec.CertManInssuer != "" {
		ing.Annotations = make(map[string]string)
		ing.Annotations["acme.cert-manager.io/http01-edit-in-place"] = "true"
		ing.Annotations["cert-manager.io/issuer"] = clientResource.Spec.CertManInssuer
	}
	if clientResource.Spec.Path != "" && clientResource.Spec.Path != "/" {
		if len(ing.Annotations) == 0 {
			ing.Annotations = make(map[string]string)
		}
		ing.Annotations["nginx.ingress.kubernetes.io/rewrite-target"] = "/$2"
	}

	pfrx := netv1.PathTypePrefix

	p := "/"
	if clientResource.Spec.Path != "" {
		p = clientResource.Spec.Path + "(/|$)(.*)"
	}

	ingressPath := netv1.HTTPIngressPath{
		Path:     p,
		PathType: &pfrx,
		Backend: netv1.IngressBackend{
			Service: &netv1.IngressServiceBackend{
				Name: serviceName,
				Port: netv1.ServiceBackendPort{
					Number: int32(clientResource.Spec.Port),
				},
			},
		},
	}
	httpIngressRuleValue := netv1.HTTPIngressRuleValue{Paths: []netv1.HTTPIngressPath{ingressPath}}
	rule := netv1.IngressRule{}
	rule.Host = clientResource.Spec.Host
	rule.IngressRuleValue.HTTP = &httpIngressRuleValue

	if clientResource.Spec.CertManInssuer != "" {
		tls := netv1.IngressTLS{
			Hosts:      []string{clientResource.Spec.Host},
			SecretName: strings.ReplaceAll(clientResource.Spec.Host, ".", "-") + "-tls",
		}
		ing.Spec = netv1.IngressSpec{
			Rules: []netv1.IngressRule{rule},
			TLS:   []netv1.IngressTLS{tls},
		}
	} else {
		ing.Spec = netv1.IngressSpec{
			Rules: []netv1.IngressRule{rule},
		}

	}

	return &ing
}

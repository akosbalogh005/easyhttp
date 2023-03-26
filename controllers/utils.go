package controllers

import (
	corev1 "k8s.io/api/core/v1"
)

func convertEnv(m map[string]string) []corev1.EnvVar {
	var ret []corev1.EnvVar
	for k, v := range m {
		ret = append(ret, corev1.EnvVar{Name: k, Value: v})
	}
	return ret
}

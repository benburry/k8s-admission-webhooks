package handlers

import (
	"encoding/json"
	"fmt"

	"k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/prometheus/prometheus/promql"
)

// Perform the admission logic on a prometheus-rules ConfigMap

type PrometheusRulesAdmissionController struct{}

func (g *PrometheusRulesAdmissionController) Admit(ar *v1beta1.AdmissionReview) error {
	configmap, err := extractConfigMap(ar)
	if err != nil {
		return err
	}

	return admitConfigMap(configmap)
}

func lintString(content string) error {
	_, err := promql.ParseStmts(content)
	return err
}

func admitConfigMap(configmap *corev1.ConfigMap) error {
	if configmap.Labels["role"] == "prometheus-rulefiles" {
		for _, content := range configmap.Data {
			if err := lintString(content); err != nil {
				return fmt.Errorf("Prometheus linter failed for ConfigMap '%s': %v", configmap.Name, err)
			}
		}
	}
	return nil
}

// fetch the ConfigMap object out of the AdmissionReview
func extractConfigMap(ar *v1beta1.AdmissionReview) (*corev1.ConfigMap, error) {
	// verify that we received a ConfigMap object
	resource := metav1.GroupVersionResource{Group: "", Version: "v1", Resource: "configmaps"}
	if ar.Request.Resource != resource {
		return nil, fmt.Errorf("expect resource to be %s", resource)
	}

	configmap := corev1.ConfigMap{}
	err := json.Unmarshal(ar.Request.Object.Raw, &configmap)
	return &configmap, err
}

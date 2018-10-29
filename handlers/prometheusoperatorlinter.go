/*    Copyright 2018 Ben Burry

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

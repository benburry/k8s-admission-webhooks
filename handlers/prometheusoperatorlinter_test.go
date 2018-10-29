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
	"testing"

	"k8s.io/api/admission/v1beta1"
)

var (
	promArJsonFmt    string = `{"kind":"AdmissionReview","apiVersion":"admission.k8s.io/v1beta1","request":{"uid":"d7e11614-4512-11e8-8d4f-b827ebf9752a","kind":{"group":"","version":"v1","kind":"ConfigMap"},"resource":{"group":"","version":"v1","resource":"configmaps"},"namespace":"default","operation":"CREATE","userInfo":{"username":"kubernetes-admin","groups":["system:masters","system:authenticated"]},"object":{"metadata":{"name":"test-rules","namespace":"default","uid":"d7e0f812-4512-11e8-8d4f-b827ebf9752a","creationTimestamp":"2018-04-21T03:19:46Z","labels":%s},"data":{"test.rules":%s}},"oldObject":null}}`
	validQuery   string = `ALERT deployment_health IF kube_deployment_status_replicas_available{deployment="test-deployment",namespace="default"} < kube_deployment_spec_replicas FOR 10m LABELS { team = "test-team", urgent = "false" } ANNOTATIONS { summary = "Deployment is unhealthy", description = "Fewer than the expected number of pods are running.", runbook = "", }`
	invalidQuery string = ` ALERT deployment_health LABELS { team = "test-team", urgent = "false" }`
)

func UnmarshalCustomLabelsAR(t *testing.T, rules string, labels string) *v1beta1.AdmissionReview {
	ar := v1beta1.AdmissionReview{}
	jsonrules, _ := json.Marshal(rules)
	if err := json.Unmarshal([]byte(fmt.Sprintf(promArJsonFmt, labels, jsonrules)), &ar); err != nil {
		t.Errorf("%+v", err)
	}
	return &ar
}

func UnmarshalPromAR(t *testing.T, rules string) *v1beta1.AdmissionReview {
	return UnmarshalCustomLabelsAR(t, rules, `{"prometheus":"shared","role":"prometheus-rulefiles"}`)
}

func TestValidSyntaxAR(t *testing.T) {
	ar := UnmarshalPromAR(t, validQuery)
	handler := PrometheusRulesAdmissionController{}
	if err := handler.Admit(ar); err != nil {
		t.Error("Expecting valid promql to be allowed")
	}
}

func TestInvalidSyntaxAR(t *testing.T) {
	ar := UnmarshalPromAR(t, invalidQuery)
	handler := PrometheusRulesAdmissionController{}
	if err := handler.Admit(ar); err == nil {
		t.Error("Expecting invalid promql to be disallowed")
	}
}

func TestInvalidSyntaxARShouldBeIgnored(t *testing.T) {
	ar := UnmarshalCustomLabelsAR(t, invalidQuery, `{"name": "test"}`)
	handler := PrometheusRulesAdmissionController{}
	if err := handler.Admit(ar); err != nil {
		t.Error("Expecting invalid promql in a non-prometheus configmap to be ignored and allowed")
	}
}

func TestValidSyntax(t *testing.T) {
	if err := lintString(validQuery); err != nil {
		t.Error("Expecting valid promql to be allowed")
	}
}

func TestInvalidSyntax(t *testing.T) {
	if err := lintString(invalidQuery); err == nil {
		t.Error("Expecting invalid promql to be disallowed")
	}
}

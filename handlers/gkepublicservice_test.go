package handlers

import (
	"encoding/json"
	"fmt"
	"testing"

	"k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
)

var arJsonFmt string = `{"kind":"AdmissionReview","request":{"kind":{"kind":"Service","version":"v1","group":""},"resource":{"group":"","version":"v1","resource":"services"},"uid":"681f0022-306f-11e8-8d4f-b827ebf9752a","object":{"status":{"loadBalancer":{}},"spec":{"clusterIP":"192.168.1.254","selector":{"name":"test-invalid-selector"},"type":"LoadBalancer","ports":[{"protocol":"TCP","targetPort":1024,"port":1024}],"sessionAffinity":"None"},"metadata":{"uid":"681ee709-306f-11e8-8d4f-b827ebf9752a","namespace":"default","creationTimestamp":"2018-03-25T20:59:28Z","annotations":{%s},"name":"test-service"}},"namespace":"default","userInfo":{"username":"kubernetes-admin","groups":["system:masters","system:authenticated"]},"oldObject":null,"operation":"CREATE"},"apiVersion":"admission.k8s.io/v1beta1"}`

var unannotatedJson string = fmt.Sprintf(arJsonFmt, "")
var annotatedJson string = fmt.Sprintf(arJsonFmt, `"cloud.google.com/load-balancer-type":"internal"`)

func UnmarshalAR(arJson string) *v1beta1.AdmissionReview {
	ar := v1beta1.AdmissionReview{}
	json.Unmarshal([]byte(arJson), &ar)
	return &ar
}

func UnmarshalService(json string) *corev1.Service {
	ar := UnmarshalAR(json)
	service, _ := extractService(ar)
	return service
}

func TestDisallowUnannotatedAR(t *testing.T) {
	ar := UnmarshalAR(unannotatedJson)

	handler := GkeServiceAdmissionController{}
	if err := handler.Admit(ar); err == nil {
		t.Error("Expecting un-annotated Service to be disallowed")
	}
}

func TestAllowAnnotatedAR(t *testing.T) {
	ar := UnmarshalAR(annotatedJson)

	handler := GkeServiceAdmissionController{}
	if err := handler.Admit(ar); err != nil {
		t.Error("Expecting annotated Service to be allowed")
	}
}

func TestDisallowExternalService(t *testing.T) {
	service := UnmarshalService(unannotatedJson)

	if err := admitService(service); err == nil {
		t.Error("Expecting un-annotated Service to be disallowed")
	}
}

func TestAllowInternalService(t *testing.T) {
	service := UnmarshalService(unannotatedJson)
	service.Annotations["cloud.google.com/load-balancer-type"] = "internal"

	if err := admitService(service); err != nil {
		t.Error("Expecting annotated Service to be allowed")
	}
}

func TestAllowNonLBService(t *testing.T) {
	service := UnmarshalService(unannotatedJson)
	service.Annotations = nil
	service.Spec.Type = "ClusterIP"

	if err := admitService(service); err != nil {
		t.Error("Expecting ClusterIP Service to be allowed")
	}
}

func TestAllowExternalServiceWithOurAnnotation(t *testing.T) {
	service := UnmarshalService(unannotatedJson)
	service.Annotations["gke/load-balancer-type"] = "external"

	if err := admitService(service); err != nil {
		t.Error("Expecting annotated Service to be allowed")
	}
}

func TestAllowExternalServiceWithBothAnnotations(t *testing.T) {
	service := UnmarshalService(unannotatedJson)
	service.Annotations["cloud.google.com/load-balancer-type"] = "external"
	service.Annotations["gke/load-balancer-type"] = "external"

	if err := admitService(service); err != nil {
		t.Error("Expecting annotated Service to be allowed")
	}
}

func TestDisallowExternalServiceWithSwappedAnnotations(t *testing.T) {
	service := UnmarshalService(unannotatedJson)
	service.Annotations["cloud.google.com/load-balancer-type"] = "external"
	service.Annotations["gke/load-balancer-type"] = "internal"

	if err := admitService(service); err == nil {
		t.Error("Expecting external Service to be disallowed")
	}
}

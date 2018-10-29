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
	"strings"

	"github.com/golang/glog"
	"k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Perform the admission logic on the Service object
//
// The rules this enforces are:
//  LoadBalancer services without annotations (that would be external by
//  default) are disallowed
//  LoadBalancer services with the google annotation marking them as internal
//  are allowed
//  LoadBalancer services with an annotation explicitly marking them as
//  intended to be visible externally are allowed
//
// The annotation that would allow an externally-visible service is:
//  gke/load-balancer-type: External
//

type GkeServiceAdmissionController struct{}

func (g *GkeServiceAdmissionController) Admit(ar *v1beta1.AdmissionReview) error {
	service, err := extractService(ar)
	if err != nil {
		return err
	}

	return admitService(service)
}

func admitService(service *corev1.Service) error {
	const annotation = "gke/load-balancer-type"

	glog.V(2).Infof("Service type %s", service.Spec.Type)

	// in GKE, only LoadBalancer services could be visible externally by default
	if service.Spec.Type == "LoadBalancer" {
		// verify that the service has the google annotation, and is set to be internal-only
		if v, found := service.Annotations["cloud.google.com/load-balancer-type"]; !(found && strings.EqualFold(v, "internal")) {
			// if it's not internal, or has no google annotation, verify that
			// our annotation is included to mark the service as external
			if v, found := service.Annotations[annotation]; !(found && strings.EqualFold(v, "external")) {
				return fmt.Errorf("The service '%s' is public, and so disallowed without the explicit '%s' annotation.", service.Name, annotation)
			}
		}
	}
	return nil
}

// fetch the actual Service object out of the AdmissionReview
func extractService(ar *v1beta1.AdmissionReview) (*corev1.Service, error) {
	// verify that we received a Service object
	serviceResource := metav1.GroupVersionResource{Group: "", Version: "v1", Resource: "services"}
	if ar.Request.Resource != serviceResource {
		return nil, fmt.Errorf("expect resource to be %s", serviceResource)
	}

	service := corev1.Service{}
	err := json.Unmarshal(ar.Request.Object.Raw, &service)
	return &service, err
}

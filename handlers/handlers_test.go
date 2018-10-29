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
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"k8s.io/api/admission/v1beta1"
)

var (
	review string = `{"kind":"AdmissionReview","apiVersion":"admission.k8s.io/v1beta1","request":{"uid":"d7e11614-4512-11e8-8d4f-b827ebf9752a","kind":{"group":"","version":"v1","kind":"ConfigMap"},"resource":{"group":"","version":"v1","resource":"configmaps"},"namespace":"default","operation":"CREATE","userInfo":{"username":"kubernetes-admin","groups":["system:masters","system:authenticated"]},"object":{"metadata":{"name":"test-rules","namespace":"default","uid":"d7e0f812-4512-11e8-8d4f-b827ebf9752a","creationTimestamp":"2018-04-21T03:19:46Z","labels":""},"data":{"test.rules":""}},"oldObject":null}}`
)

type testHandler struct {
	fail bool
}

func (h *testHandler) URL() string {
	return "/testhandler"
}

func (h *testHandler) Admit(ar *v1beta1.AdmissionReview) error {
	if h.fail {
		return errors.New("Deliberately failing test")
	}
	return nil
}

func TestContentType(t *testing.T) {
	req := httptest.NewRequest("POST", "/", strings.NewReader(review))
	w := httptest.NewRecorder()
	handleRequest(w, req, nil)

	resp := w.Result()

	if resp.StatusCode != http.StatusBadRequest {
		t.Error("Expecting non-json content type to be refused")
	}
}

func TestInvalidContent(t *testing.T) {
	req := httptest.NewRequest("POST", "/", strings.NewReader(""))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handleRequest(w, req, nil)

	resp := w.Result()

	if resp.StatusCode != http.StatusBadRequest {
		t.Error("Expecting non-ar content to be refused")
	}
}

func TestFailedAdmit(t *testing.T) {
	req := httptest.NewRequest("POST", "/", strings.NewReader(review))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handleRequest(w, req, &testHandler{fail: true})

	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Error("Expecting 200 response code")
	}

	reviewResponse := v1beta1.AdmissionReview{}
	if err := json.Unmarshal(body, &reviewResponse); err != nil {
		t.Errorf("Unable to unmarshal response: %v", err)
	}

	if reviewResponse.Response.Allowed {
		t.Error("Expecting review request to be refused")
	}
}

func TestSuccess(t *testing.T) {
	req := httptest.NewRequest("POST", "/", strings.NewReader(review))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handleRequest(w, req, &testHandler{fail: false})

	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Error("Expecting 200 response code")
	}

	reviewResponse := v1beta1.AdmissionReview{}
	if err := json.Unmarshal(body, &reviewResponse); err != nil {
		t.Errorf("Unable to unmarshal response: %v", err)
	}

	if !reviewResponse.Response.Allowed {
		t.Error("Expecting review request to be allowed")
	}
}

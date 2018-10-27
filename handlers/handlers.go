package handlers

import (
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/golang/glog"
	"k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// When writing a handler, implement this interface and call the
// RegisterHandler function below in your init{}
type AdmissionReviewHandler interface {
	// Called by Kubernetes for each object matching the `webHooks.rules` in
	// your `ValidatingWebhookConfiguration` object
	// Returning an error indicates that the admission should be rejected,
	//  with the error.Error() output displayed to the user (if the operation was
	//  initiated by a user)
	Admit(ar *v1beta1.AdmissionReview) error
}

type AdmissionReviewHandlerFuncs map[string]http.HandlerFunc

var (
	handlerFuncs AdmissionReviewHandlerFuncs
)

func init() {
	handlerFuncs = make(AdmissionReviewHandlerFuncs)
}

func RegisterHandler(url string, handler AdmissionReviewHandler) {
	handlerFuncs[url] = func(w http.ResponseWriter, r *http.Request) {
		handleRequest(w, r, handler)
	}
}

func GetRegisteredHandlers() AdmissionReviewHandlerFuncs {
	return handlerFuncs
}

func handleRequest(w http.ResponseWriter, r *http.Request, handler AdmissionReviewHandler) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		glog.Error(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	glog.V(2).Info("AdmissionReview request", string(data))

	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		glog.Errorf("contentType=%s, expect application/json", contentType)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ar := v1beta1.AdmissionReview{}
	if err := json.Unmarshal(data, &ar); err != nil {
		glog.Error(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	review := v1beta1.AdmissionReview{Response: &v1beta1.AdmissionResponse{Allowed: true, UID: ar.Request.UID}}

	if err := handler.Admit(&ar); err != nil {
		review.Response.Allowed = false
		review.Response.Result = &metav1.Status{Message: err.Error()}
	}

	resp, err := json.Marshal(review)
	if err != nil {
		glog.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(resp); err != nil {
		glog.Error(err)
	}
}

func GetServer(address string) *http.Server {
	for url, handler := range handlerFuncs {
		glog.Infof("Setting handler func for %s", url)
		http.HandleFunc(url, handler)
	}

	s := http.Server{
		Addr: address,
		TLSConfig: &tls.Config{
			ClientAuth: tls.NoClientCert,
		},
	}

	return &s
}

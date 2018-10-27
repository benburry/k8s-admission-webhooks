k8s-admission-webhooks
----

The code and configuration here serves as a framework for writing Kubernetes
[validating admission controllers][1], enforcing rules around which objects
are allowed and disallowed in the cluster.

The intention of this code is to simplify the process of adding your own
behaviour to a Kubernetes cluster, by removing as much boilerplate as possible
from writing an admission controller. Ideally this exposes only the things you
care about (processing the AdmissionReview object itself) and hides the rest of
the barriers to entry.

## Implementing your own handler
Please use the existing handlers as resources for guidance on how to implement
your own handler. They provide useful examples for extracting the specific
object you're interested in from the AdmissionReview object, amongst other
things.

The general requirement for implementing a handler are:

* Your handler must implement the `AdmissionReviewHandler` interface, defined in
  the `github.com/benburry/k8s-admission-webhooks/handlers`
  package:
  ```
  type AdmissionReviewHandler interface {
      // Called by Kubernetes for each object matching the `webHooks.rules` in
      // your `ValidatingWebhookConfiguration` object
      // Returning an error indicates that the admission should be rejected,
      //  with the error.Error() output displayed to the user (if the operation was
      //  initiated by a user)
      Admit(ar *v1beta1.AdmissionReview) error
  }
  ```

We call the `RegisterHandler` function in the
`github.com/benburry/k8s-admission-webhooks/handlers` package to
make your handler available, passing the url the handler will listen on, and
the handler function itself.


### Links
I found the following to be the most useful sources of information when
implementing these webhooks:

* https://github.com/kelseyhightower/grafeas-tutorial
* https://github.com/caesarxuchao/example-webhook-admission-controller (deprecated post k8s 1.8)
* https://github.com/kubernetes/kubernetes/tree/release-1.9/test/images/webhook

[1]: https://kubernetes.io/docs/admin/admission-controllers/#validatingadmissionwebhook-alpha-in-18-beta-in-19

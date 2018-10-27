gkepublicservice handler
----

This handler is specific to GKE, to disallow the creation of services that
are publically accessible by default. Instead, this admission controller
enforces an explicit annotation when a Service is intended to be accessible
externally.

The rules this enforces are:
* LoadBalancer services without annotations (that would be external by default on GKE) are disallowed
* LoadBalancer services with the google annotation marking them as internal are allowed
* LoadBalancer services with an annotation explicitly marking them as intended to be visible externally are allowed

To create an internal-only service, use the following Google annotation on your
Service object
```
cloud.google.com/load-balancer-type: Internal
```

To create an externally-accessible service, use the following annotation
on your Service object
```
gke/load-balancer-type: External
```


prometheus-operator linter handler
----

The prometheus-operator alert configuration is defined in ConfigMap objects,
containing rules definitions and queries. A badly-defined query can
unfortunately cause the prometheus instance to die, as it tries to load the
specified rule and fails.

This handler's purpose is to check that the syntax of these rules is correct
before they are persisted into the cluster. It does this by simply running the
`ParseStmts` function from the `github.com/prometheus/prometheus/promql`
package on the rule contents in the admit function, and passing the errors
back.

This handler assumes your prometheus-operator rules ConfigMap objects have the
following label:

```
role: prometheus-rulefiles
```

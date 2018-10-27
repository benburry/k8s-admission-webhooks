FROM alpine:latest

MAINTAINER "Ben Burry" <docker@burry.name>

EXPOSE 8000
COPY k8s-admission-webhooks /k8s-admission-webhooks
RUN mkdir -p /etc/tls

CMD ["/k8s-admission-webhooks", "-logtostderr", "-addr=:8000"]


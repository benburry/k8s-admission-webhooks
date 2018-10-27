package main

import (
	"flag"

	"github.com/golang/glog"

	"github.com/benburry/k8s-admission-webhooks/handlers"
)

func main() {
	var tlsCertFile, tlsKeyFile, addr string

	flag.StringVar(&tlsCertFile, "tls-cert", "/etc/tls/server.pem", "TLS certificate file.")
	flag.StringVar(&tlsKeyFile, "tls-key", "/etc/tls/server-key.pem", "TLS key file.")
	flag.StringVar(&addr, "addr", ":8080", "TCP address to listen on")
	flag.Parse()

	handlers.RegisterHandler("/prometheuslinter", &handlers.PrometheusRulesAdmissionController{})
	handlers.RegisterHandler("/gkepublicservice", &handlers.GkeServiceAdmissionController{})

	s := handlers.GetServer(addr)
	glog.Fatal(s.ListenAndServeTLS(tlsCertFile, tlsKeyFile))
}

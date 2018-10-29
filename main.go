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

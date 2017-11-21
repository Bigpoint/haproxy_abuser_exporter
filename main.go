/**
 * Copyright 2017 Bigpoint GmbH
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"fmt"
	"strconv"
	"flag"
	"io"
	"net/http"
	"os"
)

type config struct {
	gpc      string
	req      string
	instance string
	endpoint string
	port     int
}

type httpHandler struct {
	cfg config
}

func main() {
	gpc := flag.String("gpc", "gpc0", "The HAproxy GRC that contains the connections")
	req := flag.String("reqRate", "http_req_rate(10000)", "The HAproxy register that contains the connections")
	instance := flag.String("instance", "", "If specified will enhance the metrics with a 'instance' field")
	endpoint := flag.String("endpoint", "/metrics", "Endpoint that is exposed for prometheus")
	port := flag.Int("port", 9322, "Port to listen on")
	flag.Parse()

	cfg := config{*gpc, *req, *instance, *endpoint, *port}
	handler := httpHandler{cfg}

	http.HandleFunc(cfg.endpoint, handler.respond)
	http.ListenAndServe(fmt.Sprintf(":%d", cfg.port), nil)

	prepareOutput(cfg)
}

func (t *httpHandler) respond(w http.ResponseWriter, r *http.Request) {
	response, err := prepareOutput(t.cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		http.Error(w, "Internal Server Error", 500)
		return
	}
	io.WriteString(w, response)
}

func prepareOutput(cfg config) (string, error) {
	scraper := NewScraper("/run/haproxy/admin.sock")

	tables, err := scraper.GetTables()
	if err != nil {
		return "", fmt.Errorf("could not fetch tables")

	}

	connectedOutput := ""
	blockedOutput := ""
	reqRateOutput := ""
	output := ""
	connectedIPs := 0
	blockedIPs := 0
	instanceString := ""
	instanceString2 := ""

	if len(cfg.instance) > 0 {
		instanceString = fmt.Sprintf(",instance=\"%s\"", cfg.instance)
		instanceString2 = fmt.Sprintf("{instance=\"%s\"}", cfg.instance)
	}

	for _, table := range tables {
		tableContent, err := scraper.Scrape(table)
		if err != nil {
			return "", fmt.Errorf("could not scrape table %s", table)
		}

		for _, row := range tableContent {
			if _, ok := row["key"]; ok == false {
				continue
			}
			connectedIPs++

			if _, ok := row[cfg.gpc]; ok == true {
				connectedOutput += fmt.Sprintf("connected_ip_gpc{frontend=\"%s\",ip=\"%s\"%s} %s\n", table, row["key"], instanceString, row[cfg.gpc])

				// Parse gpc0 to an int and check it it is larger than 0
				if i, err := strconv.ParseInt(row[cfg.gpc], 10, 32); err == nil && i > 0 {
					blockedIPs ++
					// Also print the blocked message
					blockedOutput += fmt.Sprintf("blocked_ip{frontend=\"%s\",ip=\"%s\"%s} %s\n", table, row["key"], instanceString, row[cfg.gpc])
				}
			}

			if _, ok := row[cfg.req]; ok == true {
				reqRateOutput += fmt.Sprintf("http_request_rate_per_ip{frontend=\"%s\",ip=\"%s\"%s} %s\n", table, row["key"], instanceString, row[cfg.req])
			}
		}
	}

	output += "# HELP connected_ips Amount of Connected IPs\n"
	output += "# TYPE connected_ips untyped\n"
	output += fmt.Sprintf("connected_ips%s %d\n", instanceString2, connectedIPs)
	output += "# HELP blocked_ip Amount of currently blocked IPs\n"
	output += "# TYPE blocked_ips untyped\n"
	output += fmt.Sprintf("blocked_ips%s %d\n", instanceString2, blockedIPs)
	output += "# HELP connected_ip_gpc currently connected_ip gpc_counter\n"
	output += "# TYPE connected_ip_gpc gauge\n"
	output += connectedOutput
	output += "# HELP http_request_rate_per_ip currently connected_ip http_request_rate\n"
	output += "# TYPE http_request_rate_per_ip gauge\n"
	output += reqRateOutput
	output += "# HELP blocked_ip Currently blocked IPs\n"
	output += "# TYPE blocked_ip gauge\n"
	output += blockedOutput

	return output, nil
}

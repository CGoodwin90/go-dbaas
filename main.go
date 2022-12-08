package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

type funcType func() map[string]string

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/{mariadb:mariadb-.*}", mariadbHandler)
	r.HandleFunc("/{postgres:postgres-.*}", postgresHandler)
	r.HandleFunc("/{redis:redis-.*}", redisHandler)
	r.HandleFunc("/{solr:solr-.*}", solrHandler)
	r.HandleFunc("/mongo-4", mongoHandler)
	r.HandleFunc("/opensearch-2", opensearchHandler)
	r.HandleFunc("/", handleReq)
	http.Handle("/", r)
	log.Fatal(http.ListenAndServe(":3000", nil))
}

func handleReq(w http.ResponseWriter, r *http.Request) {
	var funcToCall []funcType
	for _, conFunc := range funcToCall {
		fmt.Fprintf(w, dbConnectorPairs(conFunc(), ""))
	}
}

func dbConnectorPairs(m map[string]string, connectorHost string) string {
	b := new(bytes.Buffer)
	for key, value := range m {
		fmt.Fprintf(b, "\"%s=%s\"\n", key, value)
	}
	host := fmt.Sprintf(`"SERVICE_HOST=%s"`, connectorHost)
	connectorOutput := host + "\n" + b.String()
	return connectorOutput
}

func connectorKeyValues(values []string) string {
	b := new(bytes.Buffer)
	for _, value := range values {
		if value != "" {
			v := strings.SplitN(value, ":", 2)
			fmt.Fprintf(b, "\"%s=%s\"\n", v[0], v[1])
		}
	}
	return b.String()
}

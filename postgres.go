package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

var (
	postgresUser            = os.Getenv("POSTGRES_USERNAME")
	postgresPassword        = os.Getenv("POSTGRES_PASSWORD")
	postgresDB              = os.Getenv("POSTGRES_DATABASE")
	postgresHost            = os.Getenv("POSTGRES_HOST")
	postgres11              = "postgres-11"
	postgres12              = "postgres-12"
	postgres13              = "postgres-13"
	postgresSSL             = "disable"
	postgresConnectionStr   = fmt.Sprintf("user=%s password=%s dbname=%s sslmode=%s host=%s", postgresUser, postgresPassword, postgresDB, postgresSSL, postgresHost)
	postgres11ConnectionStr = fmt.Sprintf("user=%s password=%s dbname=%s sslmode=%s host=%s", postgresUser, postgresPassword, postgresDB, postgresSSL, postgres11)
	postgres12ConnectionStr = fmt.Sprintf("user=%s password=%s dbname=%s sslmode=%s host=%s", postgresUser, postgresPassword, postgresDB, postgresSSL, postgres12)
	postgres13ConnectionStr = fmt.Sprintf("user=%s password=%s dbname=%s sslmode=%s host=%s", postgresUser, postgresPassword, postgresDB, postgresSSL, postgres13)
	postgresVersion         string
)

func postgresHandler(w http.ResponseWriter, r *http.Request) {
	time.Sleep(1 * time.Second)
	postgresRoute := r.URL.Path
	switch postgresRoute {
	case "/postgres":
		fmt.Fprintf(w, dbConnectorPairs(postgresDBConnector(postgresConnectionStr), postgresVersion))
	case "/postgres-11":
		fmt.Fprintf(w, dbConnectorPairs(postgresDBConnector(postgres11ConnectionStr), postgresVersion))
	case "/postgres-12":
		fmt.Fprintf(w, dbConnectorPairs(postgresDBConnector(postgres12ConnectionStr), postgresVersion))
	case "/postgres-13":
		fmt.Fprintf(w, dbConnectorPairs(postgresDBConnector(postgres13ConnectionStr), postgresVersion))
	}
}

func postgresDBConnector(connectionString string) map[string]string {
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		log.Print(err)
	}

	defer db.Close()

	createTable := "CREATE TABLE IF NOT EXISTS env(env_key text, env_value text)"
	_, err = db.Exec(createTable)
	if err != nil {
		panic(err.Error())
	}

	query := "INSERT INTO env(env_key, env_value) VALUES ($1, $2)"
	for _, e := range os.Environ() {

		pair := strings.SplitN(e, "=", 2)
		_, err := db.Exec(query, pair[0], pair[1])
		if err != nil {
			panic(err.Error())
		}
	}

	gitSHA := "LAGOON_%"
	rows, err := db.Query(`SELECT * FROM env where env_key LIKE $1`, gitSHA)
	if err != nil {
		log.Print(err)
	}

	db.QueryRow("SELECT VERSION()").Scan(&postgresVersion)

	defer rows.Close()
	results := make(map[string]string)
	for rows.Next() {
		var envKey, envValue string
		_ = rows.Scan(&envKey, &envValue)
		results[envKey] = envValue
	}

	return results
}

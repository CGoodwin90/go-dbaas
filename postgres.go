package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	_ "github.com/lib/pq"
)

var (
	postgresUser          = os.Getenv("POSTGRES_USERNAME")
	postgresPassword      = os.Getenv("POSTGRES_PASSWORD")
	postgresDB            = os.Getenv("POSTGRES_DATABASE")
	postgresSSL           = "disable"
	postgresVersion       string
	postgresConnectionStr string
)

func postgresHandler(w http.ResponseWriter, r *http.Request) {
	postgresPath := r.URL.Path
	localRoute, lagoonRoute := cleanRoute(postgresPath)
	lagoonUsername := os.Getenv(fmt.Sprintf("%s_USERNAME", lagoonRoute))
	lagoonPassword := os.Getenv(fmt.Sprintf("%s_PASSWORD", lagoonRoute))
	lagoonDatabase := os.Getenv(fmt.Sprintf("%s_DATABASE", lagoonRoute))
	lagoonHost := os.Getenv(fmt.Sprintf("%s_HOST", lagoonRoute))

	if localCheck != "" {
		postgresConnectionStr = fmt.Sprintf("user=%s password=%s dbname=%s sslmode=%s host=%s", lagoonUsername, lagoonPassword, lagoonDatabase, postgresSSL, lagoonHost)
		fmt.Println(postgresConnectionStr)
	} else {
		postgresConnectionStr = fmt.Sprintf("user=%s password=%s dbname=%s sslmode=%s host=%s", postgresUser, postgresPassword, postgresDB, postgresSSL, localRoute)
	}
	fmt.Fprintf(w, dbConnectorPairs(postgresDBConnector(postgresConnectionStr), postgresVersion))
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
		log.Print(err)
	}

	query := "INSERT INTO env(env_key, env_value) VALUES ($1, $2)"
	for _, e := range os.Environ() {

		pair := strings.SplitN(e, "=", 2)
		_, err := db.Exec(query, pair[0], pair[1])
		if err != nil {
			log.Print(err)
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

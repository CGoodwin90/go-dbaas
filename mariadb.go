package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

var (
	mariadbUser     = os.Getenv("MARIADB_USERNAME")
	mariadbPassword = os.Getenv("MARIADB_PASSWORD")
	mariadb         = os.Getenv("MARIADB_DATABASE")
	mariadbPort     = 3306
	mariadbVersion  string
)

func mariadbHandler(w http.ResponseWriter, r *http.Request) {
	mariadbPath := r.URL.Path
	cleanRoute := strings.ReplaceAll(mariadbPath, "/", "")
	mariadbRoute := strings.ReplaceAll(cleanRoute, "10.", "10-")
	for _, e := range os.Environ() {
		pair := strings.SplitN(e, "=", 2)
		dynamicHost := fmt.Sprintf("%s_HOST", cleanRoute)
		if pair[0] == dynamicHost {
			mariadbRoute = pair[0]
		}
	}
	mariadbConnectionStr := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", mariadbUser, mariadbPassword, mariadbRoute, mariadbPort, mariadb)
	fmt.Fprintf(w, dbConnectorPairs(mariadbConnector(mariadbConnectionStr), mariadbVersion))
}

func mariadbConnector(connectionString string) map[string]string {
	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		log.Print(err)
	}

	defer db.Close()

	createTable := "CREATE TABLE IF NOT EXISTS env(env_key text, env_value text)"
	_, err = db.Exec(createTable)
	if err != nil {
		log.Print(err)
	}

	query := "INSERT INTO env(env_key, env_value) VALUES (?, ?)"

	for _, e := range os.Environ() {
		pair := strings.SplitN(e, "=", 2)
		_, err := db.Exec(query, pair[0], pair[1])
		if err != nil {
			log.Print(err)
		}
	}

	q := "LAGOON_%"
	rows, err := db.Query(`SELECT * FROM env where env_key LIKE ?`, q)
	if err != nil {
		log.Print(err)
	}

	db.QueryRow("SELECT VERSION()").Scan(&mariadbVersion)

	defer rows.Close()
	results := make(map[string]string)
	for rows.Next() {
		var envKey, envValue string
		_ = rows.Scan(&envKey, &envValue)
		results[envKey] = envValue
	}

	return results
}

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
	mariadbUser             = os.Getenv("MARIADB_USERNAME")
	mariadbPassword         = os.Getenv("MARIADB_PASSWORD")
	mariadb                 = os.Getenv("MARIADB_DATABASE")
	mariadbHost             = os.Getenv("MARIADB_HOST")
	mariadb104              = "mariadb-10.4"
	mariadb105              = "mariadb-10.5"
	mariadbPort             = 3306
	mariadbConnectionStr    = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", mariadbUser, mariadbPassword, mariadbHost, mariadbPort, mariadb)
	mariadb104ConnectionStr = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", mariadbUser, mariadbPassword, mariadb104, mariadbPort, mariadb)
	mariadb105ConnectionStr = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", mariadbUser, mariadbPassword, mariadb105, mariadbPort, mariadb)
	mariadbVersion          string
)

func mariadbHandler(w http.ResponseWriter, r *http.Request) {
	mariadbRoute := r.URL.Path
	switch mariadbRoute {
	case "/mariadb":
		fmt.Fprintf(w, dbConnectorPairs(mariadbConnector(mariadbConnectionStr), mariadbVersion))
	case "/mariadb-10.4":
		fmt.Fprintf(w, dbConnectorPairs(mariadbConnector(mariadb104ConnectionStr), mariadbVersion))
	case "/mariadb-10.5":
		fmt.Fprintf(w, dbConnectorPairs(mariadbConnector(mariadb105ConnectionStr), mariadbVersion))
	default:
		break
	}
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
		panic(err.Error())
	}

	query := "INSERT INTO env(env_key, env_value) VALUES (?, ?)"

	for _, e := range os.Environ() {
		pair := strings.SplitN(e, "=", 2)
		_, err := db.Exec(query, pair[0], pair[1])
		if err != nil {
			panic(err.Error())
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

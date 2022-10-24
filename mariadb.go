package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

var (
	mariaUser          = os.Getenv("MARIADB_USER")
	mariaPassword      = os.Getenv("MARIADB_PASSWORD")
	mariaDB            = os.Getenv("MARIADB_DATABASE")
	mariaHost          = os.Getenv("MARIADB_HOST")
	mariaPort          = 3306
	mariaDriver        = "mysql"
	mariaConnectionStr = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", mariaUser, mariaPassword, mariaHost, mariaPort, mariaDB)
)

func mariaHandler(w http.ResponseWriter, r *http.Request) {
	resp := mariaDBConnector()
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		log.Fatalf("Error happened in JSON marshal. Err: %s", err)
	}
	w.Write(jsonResp)
}

func mariaDBConnector() map[string]string {
	db, err := sql.Open(mariaDriver, mariaConnectionStr)
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
	// mdbQuery := "MARIADB%"
	gitSHA := "LAGOON_%"
	rows, err := db.Query(`SELECT * FROM env where env_key LIKE ?`, gitSHA)
	if err != nil {
		log.Print(err)
	}

	defer rows.Close()
	results := make(map[string]string)
	for rows.Next() {
		var envKey, envValue string
		_ = rows.Scan(&envKey, &envValue)
		results[envKey] = envValue
	}

	return results
}

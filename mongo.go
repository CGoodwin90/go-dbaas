package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	mongoPort          = 27017
	mongoVersion       string
	mongoConnectionStr string
	database           string
)

func mongoHandler(w http.ResponseWriter, r *http.Request) {
	mongoPath := r.URL.Path
	localRoute, lagoonRoute := cleanRoute(mongoPath)
	lagoonUsername := os.Getenv(fmt.Sprintf("%s_USERNAME", lagoonRoute))
	lagoonPassword := os.Getenv(fmt.Sprintf("%s_PASSWORD", lagoonRoute))
	lagoonDatabase := os.Getenv(fmt.Sprintf("%s_DATABASE", lagoonRoute))
	lagoonPort := os.Getenv(fmt.Sprintf("%s_PORT", lagoonRoute))
	lagoonHost := os.Getenv(fmt.Sprintf("%s_HOST", lagoonRoute))

	if localCheck != "" {
		mongoConnectionStr = fmt.Sprintf("mongodb://%s:%s@%s:%s/%s", lagoonUsername, lagoonPassword, lagoonHost, lagoonPort, lagoonDatabase)
		database = lagoonDatabase
	} else {
		mongoConnectionStr = fmt.Sprintf("mongodb://%s:%d", localRoute, mongoPort)
		database = os.Getenv("MONGO_DATABASE")
	}
	fmt.Fprintf(w, mongoConnector(mongoConnectionStr, database))
}

func cleanMongoOutput(docs []primitive.M) string {
	valStr := fmt.Sprint(docs)
	r := regexp.MustCompile(`(?:LAGOON_\w*)\s\w*:(?:\w*)`)
	matches := r.FindAllString(valStr, -1)
	var mongoResults []string
	for _, str := range matches {
		mongoVals := strings.ReplaceAll(str, "value:", "")
		mongoResults = append(mongoResults, mongoVals)
	}

	b := new(bytes.Buffer)
	for _, value := range mongoResults {
		v := strings.SplitN(value, " ", 2)
		fmt.Fprintf(b, "\"%s=%s\"\n", v[0], v[1])
	}
	host := fmt.Sprintf(`"SERVICE_HOST=%s"`, mongoVersion)
	mongoOutput := host + "\n" + b.String()
	return mongoOutput
}

func mongoConnector(connectionString string, database string) string {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(connectionString))
	if err != nil {
		log.Print(err)
	}

	envCollection := client.Database(database).Collection("env-vars")

	deleteFilter := bson.D{{}}
	_, err = envCollection.DeleteMany(context.TODO(), deleteFilter)
	if err != nil {
		log.Print(err)
	}

	environmentVariables := []interface{}{}

	for _, e := range os.Environ() {
		pair := strings.SplitN(e, "=", 2)
		bsonData := bson.D{{"Key", pair[0]}, {"value", pair[1]}}
		environmentVariables = append(environmentVariables, bsonData)
		if err != nil {
			log.Print(err)
		}
	}

	var commandResult bson.M
	command := bson.D{{"serverStatus", 1}}
	_ = client.Database(database).RunCommand(context.TODO(), command).Decode(&commandResult)

	mongoVersion = fmt.Sprintf("Mongo:%+v", commandResult["version"])

	_, err = envCollection.InsertMany(context.TODO(), environmentVariables)
	if err != nil {
		log.Print(err)
	}
	filter := bson.D{{"Key", primitive.Regex{Pattern: "LAGOON", Options: ""}}}
	cursor, _ := envCollection.Find(context.TODO(), filter, options.Find().SetProjection(bson.M{"_id": 0}))
	var docs []bson.M
	if err = cursor.All(context.TODO(), &docs); err != nil {
		log.Print(err)
	}

	mongoOutput := cleanMongoOutput(docs)
	return mongoOutput
}

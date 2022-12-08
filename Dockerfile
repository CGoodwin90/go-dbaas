FROM golang:alpine

WORKDIR /go-dbaas

ADD . .

RUN go get github.com/joho/godotenv

RUN go build && chmod +x ./go-dbaas

ENV MARIADB_PASSWORD=lagoon \
MARIADB_USERNAME=lagoon \
MARIADB_DATABASE=lagoon \
MARIADB_HOST=mariadb-10-4 \
POSTGRES_USERNAME=lagoon \
POSTGRES_PASSWORD=lagoon \
POSTGRES_DATABASE=lagoon \
POSTGRES_HOST=postgres-14 \
SOLR_HOST=solr-8 \
REDIS_HOST=redis-6 \
OPENSEARCH_HOST=opensearch-2 \
MONGO_HOST=mongo-4 \
MONGO_DATABASE=lagoon 

EXPOSE 3000

CMD sleep 10 && ./go-dbaas
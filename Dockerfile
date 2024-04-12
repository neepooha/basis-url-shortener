# syntax=docker/dockerfile:1
ARG GO_VERSION=1.22.1
FROM golang:${GO_VERSION}-alpine AS base
WORKDIR /apps/url-shortener

COPY go.mod go.sum ./
RUN go mod download
COPY . .

FROM base as build-migrate
CMD [ "go run ./cmd/migrator/ --url=myuser:mypass@urldb:5432 --dbname=url --migrations-path=./migrations"] 

FROM base as build-restapi
RUN go build -o url-shortener ./cmd/url-shortener/main.go
EXPOSE 8080
ENTRYPOINT [ "./url-shortener" ]
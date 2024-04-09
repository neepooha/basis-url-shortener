# syntax=docker/dockerfile:1
ARG GO_VERSION=1.22.1
FROM golang:${GO_VERSION}-alpine AS base
WORKDIR /apps/url-shortener

COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o url-shortener ./cmd/url-shortener/main.go
ENTRYPOINT [ "./url-shortener" ]

FROM base AS build-migrator
RUN go build -o migrations ./cmd/migrator/main.go
RUN ./migrations

EXPOSE 8080
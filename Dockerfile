# syntax=docker/dockerfile:1
ARG GO_VERSION=1.22.1
FROM golang:${GO_VERSION}-alpine AS base
WORKDIR /apps/url-shortener

RUN --mount=type=cache,target=/go/pkg/mod/ \
--mount=type=bind,source=go.sum,target=go.sum \
--mount=type=bind,source=go.mod,target=go.mod \
go mod download -x

FROM base AS build-restapi
RUN --mount=type=cache,target=/go/pkg/mod/ \
--mount=type=bind,target=. \
go build -o /url-shortener ./cmd/url-shortener/main.go

FROM scratch AS server
COPY ./config.env ./
COPY ./config ./config
COPY ./migrations ./migrations
COPY --from=build-restapi /url-shortener ./
EXPOSE 8080
ENTRYPOINT [ "./url-shortener" ]
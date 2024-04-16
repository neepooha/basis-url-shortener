<div align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="https://github.com/neepooha/url-shortener/main/raw/assets/images/logo-white.png">
    <img alt="Daytona logo" src="https://github.com/neepooha/url-shortener/main/raw/assets/images/logo-black.png" width="40%">
  </picture>
  <br>
  Url shortener
  <br>
</div>

<br><br>

<div align="center">
[![License](https://img.shields.io/badge/License-MIT-red)](#license)
[![Go Report Card](https://goreportcard.com/badge/github.com/neepooha/url_shortener)](https://goreportcard.com/report/github.com/neepooha/url_shortener)
[![Go Reference](https://pkg.go.dev/badge/github.com/neepooha/url_shortener.svg)](https://pkg.go.dev/github.com/neepooha/url_shortener)
[![Build Status](https://github.com/neepooha/url_shortener/actions/workflows/deploy.yml/badge.svg)](https://github.com/neepooha/url_shortener/actions/workflows/deploy.yml)
<br>
</div>

## Features

This project was implemented for the purpose of learning, knowledge [SOLID principles](https://en.wikipedia.org/wiki/SOLID) 
and [clean architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html).
It encourages writing clean and idiomatic Go code.

The project used:

* RESTful endpoints in the widely accepted format
* Standard CRUD operations of a database table
* JWT-based authentication
* Environment dependent application configuration management
* Structured logging with contextual information
* Error handling with proper error response generation
* Database migration
* Data validation
* Containerizing an application in Docker
 
The kit uses the following Go packages:

* Routing: [go-chi](https://github.com/go-chi/chi)
* Database access: [pgx](https://github.com/jackc/pgx)
* Database migration: [golang-migrate](https://github.com/golang-migrate/migrate)
* Data validation: [go-playground validator](https://github.com/go-playground/validator)
* Logging: [log/slog](https://pkg.go.dev/golang.org/x/exp/slog)
* JWT: [jwt-go](https://github.com/dgrijalva/jwt-go)
* Config reader: [cleanv](github.com/ilyakaznacheev/cleanenv)  
* Env reader: [godotenv](github.com/joho/godotenv)

<div align="center">
This project is a microservice that works in conjunction with a SSO microservice. For full functionality, you need to have two microservices running.
[![SSO](https://github-readme-stats.vercel.app/api/pin/?username=neepooha&repo=sso&border_color=7F3FBF&bg_color=0D1117&title_color=C9D1D9&text_color=8B949E&icon_color=7F3FBF)](https://github.com/neepooha/sso)
<br>
</div>

## Getting Started

If this is your first time encountering Go, please follow [the instructions](https://golang.org/doc/install) to install Go on your computer. 
The project requires Go 1.21 or above.

[Docker](https://www.docker.com/get-started) is also needed if you want to try the kit without setting up your own database server.
The project requires Docker 17.05 or higher for the multi-stage build support.

Also for simple run commands i use [Taskfile](https://taskfile.dev/installation/). 

After installing Go, Docker and TaskFile, run the following commands to start experiencing:
```shell
## RUN URL-SHORTENER
# download the project
git clone https://github.com/neepooha/url_shortener.git
cd url_shortener

# create config.env with that text:
> ./config.env {
CONFIG_PATH=./config/local.yaml
POSTGRES_DB=url
POSTGRES_USER=myuser
POSTGRES_PASSWORD=mypass
}

# start a PostgreSQL database server in a Docker container
task db-start

# run the RESTful API server
go run ./cmd/url-shortener

## RUN SSO
# download the project
git clone [https://github.com/neepooha/url_shortener.git](https://github.com/neepooha/sso)
cd sso

# create config.env with that text:
> ./config.env {
CONFIG_PATH=./config/local.yaml
POSTGRES_DB=url
POSTGRES_USER=myuser
POSTGRES_PASSWORD=mypass
}

# start a PostgreSQL database server in a Docker container
task db-start

# run the SSO server
go run ./cmd/sso
```
Also, you can start project in dev mode. For that you need rename in config.env
"CONFIG_PATH=./config/local.yaml" to "CONFIG_PATH=./config/dev.yaml" in both projects
and run following commads:
```shell
# run the RESTful API server with docker-compose
cd url_shortener/
docker compose up --build

# run the SSO server
cd sso/
docker compose up --build
```

At this time, you have a RESTful API server running at http://localhost:8080 and SSO-grpc Server running at http://localhost:44044.  Restful-API server provides the following endpoints:

* `POST /urls`: shortens the link using an alias, or if the alias is not specified, then using a random 6-digit cache. Need authentication
* `DELETE /urls/{alias}`: remove link by alias. You need to be an admin
* `GET /{alias}`: redirect by alias (all users)

* `POST /user`: creates a new admin. You need to be an creator
* `DELETE /user`: deletes an admin. You need to be an creator

## Project Layout
Project has the following project layout:
```
url-shortener/
├── cmd/                       start of applications of the project
├── config/                    configuration files for different environments
├── deployment/                configuration for create daemon in linux
├── internal/                  private application and library code
│   ├── app/                   application assembly
│   ├── clients/               grpc servrers(only sso) assembly
│   ├── config/                configuration library
│   ├── lib/                   additional functions for logging, error handling, migration
│   ├── storage/               storage library
│   └── transport/             handlers and middlewares
│       ├── handlers/          handlres
│       │   ├── admins/        handlres to set/delete admins
│       │   └── url/           handlres to set/delete urls
│       └── middleware/        middlewares
│           ├── auth/          middleware for auth
│           ├── context/       middleware for auth
│           ├── isadmin/       middleware for check is admin user
│           └── logger/        logger for middleware
├── migrations/                migrations
├── .gitignore
└── config.env                 config for sercret variables
```
The top level directories `cmd`, `internal`, `lib` are commonly found in other popular Go projects, as explained in
[Standard Go Project Layout](https://github.com/golang-standards/project-layout).

Within `internal` package are structured by features in order to achieve the so-called
[screaming architecture](https://blog.cleancoder.com/uncle-bob/2011/09/30/Screaming-Architecture.html). For example, 
the `transport` directory contains the application logic related with the entity feature. 

Within each feature package, code are organized in layers (API, service, repository), following the dependency guidelines
as described in the [clean architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html).

### Updating Database Schema
for simple migration you can use the following commands
```shell
# For up migrations
task up

# For drop migrations
task drop

# Revert the last database migration.
tasl rollback
```
### Managing Configurations

The application configuration is represented in `internal/config/config.go`. When the application starts,
it loads the configuration from a configuration environment as well as environment variables. The path to the configuration environment
should be in the project root folder.

The `config` directory contains the configuration files named after different environments. For example,
`config/local.yml` corresponds to the local development environment and is used when running the application 
via `go run ./cmd/url-shortener`

You can keep secrets in local/dev confing, but do not keep secrets in the prud and in the configuration environment.
For set secret variable user github secrets and deploy.yaml.

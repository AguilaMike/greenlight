# Greenlight: Supporting Material for the "Let's Go Further" Book 📖

---

## Table of Contents 📋

- [Greenlight: Supporting Material for the "Let's Go Further" Book 📖](#greenlight-supporting-material-for-the-lets-go-further-book-)
  - [Table of Contents 📋](#table-of-contents-)
  - [Installation 🛠️](#installation-️)
    - [Install dependencies](#install-dependencies)
    - [Install database](#install-database)
    - [Set environment variables](#set-environment-variables)
  - [Usage 🚀](#usage-)
  - [Project Structure 📂](#project-structure-)
    - [Endpoints](#endpoints)
  - [Prerequisites ✔️](#prerequisites-️)
  - [Contribute 🤝](#contribute-)
  - [Activities](#activities)

## Installation 🛠️

### Install dependencies

To install the code on your local machine, you need to install all the dependencies with the following command:

```go
go mod tidy
```

### Install database

Before running the project, you must create a MySQL database with Docker-compose:

```bash
docker-compose -p greenlight up -d --build
```

### Set environment variables

Create a `.env` file in the root of the project with the following content and configure your environment variables:

```bash
PORT=
ENV=
DB_DSN=
DB_MAX_OPEN_CONNS=
DB_MAX_IDLE_CONNS=
DB_MAX_IDLE_TIME=
LIMITER_RPS=
LIMITER_BURST=
LIMITER_ENABLED=
SMTP_HOST=
SMTP_PORT=
SMTP_USERNAME=
SMTP_PASSWORD=
SMTP_SENDER=
CORS_TRUSTED_ORIGINS=
```

> [!WARNING]
> You can ommite all variables and only create the file blank, the variables are optional.
> The variables ommited will set load by the flags in the application configurate in package config.

## Usage 🚀

Well, we are done installing everything. We must execute the following command to run the project.

```go
go run ./cmd/api
```

You can send application parameters if you need to configure other parameters.

- port: config API server port (-port 8080)
- env: config enviroment (development|staging|production) (-env production)

## Project Structure 📂

```
.
├── bin 🕸️
├── cmd 📂
│   └── api 🕸️
│       └── main.go 📄
├── internal 📂
│   ├── config 🕸️
│   │   └── config.go 📄
│   ├── data 📂
│   │   ├── filters.go 📄
│   │   ├── models.go 📄
│   │   ├── movies.go 📄
│   │   ├── permissions.go 📄
│   │   ├── runtime.go 📄
│   │   ├── tokens.go 📄
│   │   └── users.go 📄
│   ├── database 📂
│   │   └── db.go 📄
│   ├── mailer 📂
│   │   ├── templates 📂
│   │   │   └── user_welcome.tmpl 📄
│   │   └── mailer.go 📄
│   ├── rest 📂
│   │   ├── handlers 📂
│   │   │   ├── handlers.go 📄
│   │   │   ├── movies.go 📄
│   │   │   ├── tokens.go 📄
│   │   │   └── users.go 📄
│   │   ├── middlewares 📂
│   │   │   ├── context.go 📄
│   │   │   └── middleware.go 📄
│   │   └── routes 📂
│   │       └── routes.go 📄
│   ├── server 📂
│   │   └── server.go 📄
│   └── validator 📂
│       └── validator.go 📄
├── pkg 📂
│   └── utilities 📂
│       └── rest 📂
│           ├── handler 📂
│           │   └── handler.go 📄
│           └── helper 📂
│               ├── errors.go 📄
│               ├── helper.go 📄
│               ├── json.go 📄
│               ├── params.go 📄
│               └── worker.go 📄
├── remote 🖥️
├── scripts 📂
│   ├── migrations 📂
│   └── init.sql.go 📄
├── .env 📄
├── docker-compose.yml 📄
├── go.mod 📄
└── Makefile 📄
```

> [!NOTE]
>
> - The **bin** directory will contain our compiled application binaries, ready for deployment to a production server.
> - The **cmd/api** directory will contain the application-specific code for our Greenlight API application. This will include the code for running the server, reading and writing HTTP requests, and managing authentication.
> - The **internal** directory will contain various ancillary packages used by our API. It will contain the code for interacting with our database, doing data validation, sending emails and so on. Basically, any code which isn’t application-specific and can potentially be reused will live in here. Our Go code under cmd/api will import the packages in the internal directory (but never the other way around).
> - The **migrations** directory will contain the SQL migration files for our database.
> - The **pkg** directory will contain various shared packages used in many projects.
> - The **remote** directory will contain the configuration files and setup scripts for our production server.
> - The **go.mod** file will declare our project dependencies, versions and module path.
> - The **Makefile** will contain recipes for automating common administrative tasks — like auditing our Go code, building binaries, and executing database migrations.

### Endpoints

| Method | URL Pattern               | Required permisson    | Handler                          | Action                                 | QueryParams                          |
| :----- | :------------------------ | :-------------------- | :------------------------------- | :------------------------------------- | :----------------------------------- |
| GET    | /v1/healthcheck           | -                     | healthcheckHandler               | Show application information           |                                      |
| GET    | /v1/movies                | activate movies:read  | listMoviesHandler                | Show the details of all movies         | title, genres, page, page_size, sort |
| POST   | /v1/movies                | activate movies:write | createMovieHandler               | Create a new movie                     |                                      |
| GET    | /v1/movies/:id            | activate movies:read  | showMovieHandler                 | Show the details of a specific movie   |                                      |
| PATCH  | /v1/movies/:id            | activate movies:write | updateMovieHandler               | Update the details of a specific movie |                                      |
| DELETE | /v1/movies/:id            | activate movies:write | deleteMovieHandler               | Delete a specific movie                |                                      |
| POST   | /v1/users                 | -                     | registerUserHandler              | Register a new user                    |                                      |
| PUT    | /v1/users/activated       | -                     | activateUserHandler              | Activate a specific user               |                                      |
| POST   | /v1/tokens/authentication | -                     | createAuthenticationTokenHandler | Generate a new authentication token    |                                      |
| GET    | /debug/vars               | -                     | expvar.Handler()                 | Display application metrics            |                                      |

## Prerequisites ✔️

- [Go](https://golang.org/doc/install) (version 1.23 o lastest)

## Contribute 🤝

- Fork the project
- Create a branch for your feature (git checkout -b feature/new-feature)
- Make your changes and commit (git commit -am 'Add new feature')
- Push your changes to your fork (git push origin feature/new-feature)
- Open a Pull Request

## Activities

- [X] Creating and using Makefiles
- [X] Managing environment variables
- [ ] Quality controlling code
- [ ] Module proxies and vendoring
- [ ] Bulding binaries
- [ ] Managing and automating version numbers
- [ ] update main readme
- [ ] push to main

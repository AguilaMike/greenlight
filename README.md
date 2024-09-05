# Greenlight: Supporting Material for the "Let's Go Further" Book 📖
---

## Table of Contents 📋
- [Greenlight: Supporting Material for the "Let's Go Further" Book 📖](#greenlight-supporting-material-for-the-lets-go-further-book-)
  - [Table of Contents 📋](#table-of-contents-)
  - [Installation 🛠️](#installation-️)
    - [Install dependencies](#install-dependencies)
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
│   │   ├── movies.go 📄
│   │   └── runtime.go 📄
│   └── rest 📂
│       ├── handlers 📂
│       │   ├── handlers.go 📄
│       │   └── movies.go 📄
│       ├── middlewares 📂
│       │   └── middleware.go 📄
│       └── routes 📂
│           └── routes.go 📄
├── migrations 📂
├── pkg 📂
│   └── utilities 📂
│       └── rest 📂
│           ├── handler 📂
│           │   └── handler.go 📄
│           └── helper 📂
│               ├── errors.go 📄
│               ├── helper.go 📄
│               └── json.go 📄
├── remote 🖥️
├── go.mod 📄
└── Makefile 📄
```
> [!NOTE]
> - The **bin** directory will contain our compiled application binaries, ready for deployment to a production server.
> - The **cmd/api** directory will contain the application-specific code for our Greenlight API application. This will include the code for running the server, reading and writing HTTP requests, and managing authentication.
> - The **internal** directory will contain various ancillary packages used by our API. It will contain the code for interacting with our database, doing data validation, sending emails and so on. Basically, any code which isn’t application-specific and can potentially be reused will live in here. Our Go code under cmd/api will import the packages in the internal directory (but never the other way around).
> - The **migrations** directory will contain the SQL migration files for our database.
> - The **pkg** directory will contain various shared packages used in many projects.
> - The **remote** directory will contain the configuration files and setup scripts for our production server.
> - The **go.mod** file will declare our project dependencies, versions and module path.
> - The **Makefile** will contain recipes for automating common administrative tasks — like auditing our Go code, building binaries, and executing database migrations.

### Endpoints
| Method | URL Pattern | Handler | Action |
| :--- | :--- |  :--- |  :--- |
| GET | /v1/healthcheck | healthcheckHandler | Show application information |
| POST | /v1/movies | createMovieHandler | Create a new movie |
| GET | /v1/movies/:id | showMovieHandler | Show the details of a specific movie |

## Prerequisites ✔️

- [Go](https://golang.org/doc/install) (version 1.23 o lastest)

## Contribute 🤝

- Fork the project
- Create a branch for your feature (git checkout -b feature/new-feature)
- Make your changes and commit (git commit -am 'Add new feature')
- Push your changes to your fork (git push origin feature/new-feature)
- Open a Pull Request

## Activities

- [ ] Read and customize JSON request decoding from REST API
- [ ] Wrap errors request and send the responses
- [ ] Restrict inputs
- [ ] Validating JSON inputs
- [ ] update main readme
- [ ] push to main

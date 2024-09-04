# Greenlight: Supporting Material for the "Let's Go Further" Book 📖
---

## Table of Contents 📋
- [Greenlight: Supporting Material for the "Let's Go Further" Book 📖](#greenlight-supporting-material-for-the-lets-go-further-book-)
  - [Table of Contents 📋](#table-of-contents-)
  - [Installation 🛠️](#installation-️)
  - [Usage 🚀](#usage-)
  - [Project Structure 📂](#project-structure-)
  - [Prerequisites ✔️](#prerequisites-️)
  - [Contribute 🤝](#contribute-)
  - [Activities](#activities)

## Installation 🛠️

## Usage 🚀

## Project Structure 📂

```
.
├── bin 🕸️
├── cmd 📂
│   └── api 🕸️
│       └── main.go 📄
├── internal 📂
├── migrations 📂
├── remote 🖥️
├── go.mod 📄
└── Makefile 📄
```
> [!NOTE]
> - The **bin** directory will contain our compiled application binaries, ready for deployment to a production server.
> - The **cmd/api** directory will contain the application-specific code for our Greenlight API application. This will include the code for running the server, reading and writing HTTP requests, and managing authentication.
> - The **internal** directory will contain various ancillary packages used by our API. It will contain the code for interacting with our database, doing data validation, sending emails and so on. Basically, any code which isn’t application-specific and can potentially be reused will live in here. Our Go code under cmd/api will import the packages in the internal directory (but never the other way around).
> - The **migrations** directory will contain the SQL migration files for our database.
> - The **remote** directory will contain the configuration files and setup scripts for our production server.
> - The **go.mod** file will declare our project dependencies, versions and module path.
> - The **Makefile** will contain recipes for automating common administrative tasks — like auditing our Go code, building binaries, and executing database migrations.

## Prerequisites ✔️

- [Go](https://golang.org/doc/install) (version 1.23 o lastest)

## Contribute 🤝

- Fork the project
- Create a branch for your feature (git checkout -b feature/new-feature)
- Make your changes and commit (git commit -am 'Add new feature')
- Push your changes to your fork (git push origin feature/new-feature)
- Open a Pull Request

## Activities

- [X] Create skeleton directory structure
- [ ] Establish a HHTP server to listen fo incomming HTTP request
- [ ] add command-line flags
- [ ] add httprouter package to help implement a standar RESTful structure for the API endpoints
- [ ] update main readme
- [ ] push to main

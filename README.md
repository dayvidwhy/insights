# Insights

![GitHub last commit](https://img.shields.io/github/last-commit/dayvidwhy/insights)
![GitHub issues](https://img.shields.io/github/issues/dayvidwhy/insights)
![GitHub pull requests](https://img.shields.io/github/issues-pr/dayvidwhy/insights)
![GitHub](https://img.shields.io/github/license/dayvidwhy/insights)

Insights provides real-time analytics for web applications, without impacting site performance.

## Prerequisites

Before you begin, ensure you have the following installed:
- Docker
- Git

## Getting Started

The development environment is provided by containers.

```bash
git clone git@github.com:dayvidwhy/insights.git
cd insights
docker-compose up --build
docker exec -it insights-app bash
```

Copy the example env file and update the variables.

```bash
cp .env.example .env
```

Build and run the server:
```bash
go build -o bin/server server.go
./bin/server
```

Or with go run:

```bash
go run server.go
```

Server will be available at `localhost:1323` on your machine.

## VSCode Integration
For an optimized development experience, attach VSCode to the running insights-app container:

1. Use the command palette (Ctrl+Shift+P or Cmd+Shift+P on Mac) and select: `>Dev Containers: Attach to Running Container...`
2. Choose /insights-app from the list.

VSCode will recommend the Go extension when you open a `.go` file.

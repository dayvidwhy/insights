# Insights

Page view analytics.

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

Build and run the server.
```bash
go build -o bin/server server.go
./bin/server

# or with go run
go run server.go
```

Server will be available at `localhost:1323` on your machine.

## Development
Attach VSCode to the running `insights-app` container and install the Go extension for autocompletion. This saves needing to install Go locally and speeds up the setup.

`>Dev Containers: Attach to Running Container...`

Select `/insights-app`.

VSCode will recommend the Go extension when you open a `.go` file.

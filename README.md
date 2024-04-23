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
```

Server will be available at `localhost:1323` on your machine.

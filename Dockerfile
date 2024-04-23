FROM golang:latest

WORKDIR /app

COPY . .

EXPOSE 1323

# hold the container open
CMD ["tail", "-f", "/dev/null"]

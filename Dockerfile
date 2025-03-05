FROM golang:1.23-alpine

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY app app

RUN go build -o main app/main.go

EXPOSE 8080

# TODO: separate into builder and exec stages?

CMD ["./main"]

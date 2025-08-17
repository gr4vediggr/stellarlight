FROM golang:1.25 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

ARG name
ENV NAME=$name

COPY . .

# list the directories to be built
RUN ls -d ./cmd/*

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/run ./cmd/$name

FROM alpine:latest

WORKDIR /app

COPY ./assets ./assets
COPY --from=builder /app/bin/run ./

CMD ["./run"]
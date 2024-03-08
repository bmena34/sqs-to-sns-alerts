FROM golang:1.22-alpine AS build-base

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY *.go ./

RUN GOOS=linux GOARCH=amd64 go build -o sqs-to-sns-alerts *.go

FROM alpine

WORKDIR /app

COPY --from=build-base /sqs-to-sns-alerts /app

CMD [ "/sqs-to-sns-alerts" ]
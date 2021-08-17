FROM golang:1.16-stretch as builder

RUN apt-get update
RUN apt-get install -y git
RUN apt-get install -y build-essential
WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download
RUN go mod verify

COPY *.go ./
RUN GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /onionfiles

COPY static/*.html ./static/
COPY templates/*.html ./templates/
RUN mkdir ./static/files
RUN touch ./static/files/.empty


EXPOSE 8080

ENTRYPOINT ["/onionfiles"]


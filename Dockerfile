FROM golang:1.16-stretch as builder

RUN apt-get update
RUN apt-get install -y git
RUN apt-get install -y build-essential
WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download
# This is a workaround for the long build times libtor has
# build all of the vendor dependencies before moving on to the code
# this way if we have to rebuild the code, the vendor dependencies don't have to be rebuilt
# See https://www.reddit.com/r/golang/comments/hj4n44/improved_docker_go_module_dependency_cache_for/
RUN go list -m all | tail -n +2 | cut -f 1 -d " " | awk 'NF{print $0 "/..."}' | GOOS=linux xargs -n1 go build -v -installsuffix cgo -i; echo done

COPY *.go ./
RUN GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /onionfiles

COPY static/*.html ./static/
RUN mkdir ./keys
RUN mkdir -p ./static/files
RUN touch ./static/files/.empty
COPY static/*.html ./static/

EXPOSE 8080

ENTRYPOINT ["/onionfiles"]


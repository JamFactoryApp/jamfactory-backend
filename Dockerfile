# golang:1.15.0-alpine
FROM golang@1.14-alpine as builder

WORKDIR $GOPATH/src/github.com/jamfactoryapp/jamfactory-backend

COPY go.mod .
ENV GO111MODULE=on
RUN go mod download
RUN go mod verify

COPY . .

CMD ["go", "run", "./server.go"]

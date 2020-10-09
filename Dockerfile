FROM golang:1.14-alpine as builder

WORKDIR $GOPATH/src/github.com/jamfactoryapp/jamfactory-backend

COPY go.mod .
RUN go mod download
RUN go mod verify

COPY . .

CMD ["go", "run", "./server.go"]

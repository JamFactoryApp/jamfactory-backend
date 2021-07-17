FROM golang:1.16-alpine as builder

WORKDIR $GOPATH/src/github.com/jamfactoryapp/jamfactory-backend

COPY go.mod .
COPY go.sum .
RUN go mod download
RUN go mod verify

COPY . .

CMD ["go", "run", "./cmd/jamfactory/main.go"]

# golang:1.15.0-alpine
FROM golang@sha256:59eae48746048266891b7839f7bb9ac54a05cec6170f17ed9f4fd60b331b644b as builder

WORKDIR $GOPATH/src/jamfactory-backend

COPY go.mod .

ENV GO111MODULE=on
RUN go mod download
RUN go mod verify

COPY . .

CMD ["go", "run", "./server.go"]

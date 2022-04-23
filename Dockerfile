FROM golang:1.16-alpine as builder
WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download
RUN go mod verify
RUN go get -u github.com/cosmtrek/air
COPY ./ /app/
ENTRYPOINT ["air"]

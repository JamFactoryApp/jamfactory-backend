FROM golang:1.19-alpine as builder
WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download
RUN go mod verify
RUN go install github.com/cosmtrek/air@latest

COPY ./ /app/
CMD ["air", "-c", ".air.toml"]

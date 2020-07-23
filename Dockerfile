# golang alpine 1.14.2
FROM golang@sha256:b0678825431fd5e27a211e0d7581d5f24cede6b4d25ac1411416fa8044fa6c51 as builder

WORKDIR $GOPATH/src/jamfactory-backend

COPY go.mod .

ENV GO111MODULE=on
RUN go mod download
RUN go mod verify

COPY . .

CMD ["go", "run", "./server.go"]

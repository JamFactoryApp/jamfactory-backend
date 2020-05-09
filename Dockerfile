FROM golang:1.14.2-alpine as build

WORKDIR /go/src/jamfactory-backend

COPY go.mod .
COPY go.sum .
RUN go mod download -x

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o /go/bin/jamfactory-backend


FROM scratch
COPY --from=build /go/bin/jamfactory-backend /go/bin/jamfactory-backend
ENTRYPOINT ["/go/bin/jamfactory-backend"]

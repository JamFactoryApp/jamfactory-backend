FROM golang:1.14.2

WORKDIR /go/src/jamfactory-backend
COPY . .

RUN go get -d -v ./...
RUN go install -v ./...

CMD ["jamfactory-backend"]

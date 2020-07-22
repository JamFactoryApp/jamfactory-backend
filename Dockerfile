# golang alpine 1.14.2
FROM golang@sha256:b0678825431fd5e27a211e0d7581d5f24cede6b4d25ac1411416fa8044fa6c51 as builder

RUN apk update && apk add --no-cache git tzdata ca-certificates

ENV USER=appuser
ENV UID=10001

# See https://stackoverflow.com/a/55757473/12429735
RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/nonexistent" \
    --shell "/sbin/nologin" \
    --no-create-home \
    --uid "${UID}" \
    "${USER}"
WORKDIR $GOPATH/src/jamfactory-backend

COPY go.mod .

ENV GO111MODULE=on
RUN go mod download
RUN go mod verify

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags='-w -s -extldflags "-static"' -a -o /go/bin/jamfactory-backend .

############################
# buil small image
############################
FROM scratch

COPY --from=builder /etc/ssl/certs /etc/ssl/certs
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group

COPY --from=builder /go/bin/jamfactory-backend /go/bin/jamfactory-backend

USER appuser:appuser

ENTRYPOINT ["/go/bin/jamfactory-backend"]

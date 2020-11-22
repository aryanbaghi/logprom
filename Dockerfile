FROM golang:alpine3.12 as builder
RUN apk update && apk add git
COPY ./go.* /go/src/github.com/aryanbaghi/logprom/
WORKDIR /go/src/github.com/aryanbaghi/logprom/
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o /go/bin/logprom

FROM alpine:3.11
COPY --from=builder /go/bin/logprom /usr/bin/logprom
ENTRYPOINT [ "logprom", "serve", "-c", "/etc/logprom.yml" ]
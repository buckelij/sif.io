FROM golang:1-alpine

WORKDIR /usr/src/app
ENV CGO_ENABLED=0
COPY . .
RUN go test ./...
RUN go build -v -o /usr/local/bin/app cmd/smtp/main.go

FROM alpine:3
WORKDIR /usr/src/app
COPY --from=0 /usr/local/bin/app /usr/local/bin/app

RUN adduser -SDH sif
USER sif
CMD ["app"]

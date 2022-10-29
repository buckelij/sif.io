FROM golang:1-alpine

WORKDIR /usr/src/app
ENV CGO_ENABLED=0
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY main.go main_test.go ./
RUN go test
RUN go build -v -o /usr/local/bin/app ./...

FROM alpine:3
WORKDIR /usr/src/app
COPY --from=0 /usr/local/bin/app /usr/local/bin/app

RUN adduser -SDH sif
USER sif
CMD ["app"]

FROM golang:1-alpine

WORKDIR /usr/src/app
ENV GOARCH=arm64
ENV CGO_ENABLED=0
COPY . .
RUN go test ./...
RUN go build -v -o /usr/local/bin/app cmd/www/main.go

# sudo apt-get install qemu binfmt-support qemu-user-static
# docker run --rm --privileged multiarch/qemu-user-static --reset -p yes
FROM --platform=linux/arm64/v8 arm64v8/alpine:3
WORKDIR /usr/src/app
COPY --from=0 /usr/local/bin/app /usr/local/bin/app

RUN adduser -SDH sif
USER sif
CMD ["app"]

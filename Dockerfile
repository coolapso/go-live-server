FROM --platform=$BUILDPLATFORM golang:latest AS builder
ARG TARGETARCH
ARG VERSION


WORKDIR /go-live-server
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} go build -ldflags="-X github.com/coolapso/go-live-server/cmd.Version=${VERSION}" -a -o live-server 

FROM alpine:latest

COPY --from=builder go-live-server/live-server /usr/bin/live-server
RUN mkdir /data

EXPOSE 8080
ENTRYPOINT ["/usr/bin/live-server", "--browser", "-d", "/data"]

FROM golang:1.12.1 as builder
MAINTAINER Vijay Poliboyina
WORKDIR /root/workspace
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -mod vendor -o bin/appmeta ./cmd/server


FROM alpine:latest
COPY --from=builder /root/workspace/bin/appmeta /root/appmeta
EXPOSE 8080
ENTRYPOINT ["/root/appmeta", "-addr=:8080"]


FROM alpine:latest as root-certs
RUN apk add -U --no-cache ca-certificates
RUN addgroup -g 1001 app
RUN adduser app -u 1001 -D -G app /home/app

FROM golang:latest as builder
WORKDIR /app/
COPY --from=root-certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY . .
COPY .env .env
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod=mod -o client ./client.go

FROM scratch as final
COPY --from=root-certs /etc/passwd /etc/passwd
COPY --from=root-certs /etc/group /etc/group
COPY --chown=1001:1001 --from=root-certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --chown=1001:1001 --from=builder /app/client /app/
COPY --chown=1001:1001 --from=builder /app/.env /app/.env
USER app
EXPOSE 9090
ENTRYPOINT ["/app/client"]
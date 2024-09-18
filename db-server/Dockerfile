FROM alpine:latest as root-certs
RUN apk add -U --no-cache ca-certificates
RUN addgroup -g 1001 app
RUN adduser app -u 1001 -D -G app /home/app

FROM golang:latest as builder
WORKDIR /app/
COPY --from=root-certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY . .
COPY .env .env
COPY /docs /app/docs
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod=mod -o server ./server.go

FROM scratch as final
COPY --from=root-certs /etc/passwd /etc/passwd
COPY --from=root-certs /etc/group /etc/group
COPY --chown=1001:1001 --from=root-certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --chown=1001:1001 --from=builder /app/server /app/
COPY --chown=1001:1001 --from=builder /app/.env /app/.env
COPY --chown=1001:1001 --from=builder /app/docs /app/docs
USER app
EXPOSE 9090
ENTRYPOINT ["/app/server"]
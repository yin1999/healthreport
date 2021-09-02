FROM golang:latest as builder

WORKDIR /app

COPY . ./
RUN go mod download && go run _script/make.go

FROM debian:bullseye-slim

WORKDIR /app

COPY --from=builder /app/healthreport /app/healthreport

VOLUME ["/run/secrets"]

ENTRYPOINT ./healthreport -u="$username" -p="$password" -account="/run/secrets/account.json" -email="/run/secrets/email.json"

FROM golang:alpine as builder

WORKDIR /app

COPY . ./
# static linking
ENV CGO_ENABLED=0
RUN go mod download && go run _script/make.go

FROM busybox:latest

WORKDIR /app

COPY --from=builder /app/healthreport /app/healthreport
# add cert file
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

VOLUME ["/run/secrets"]

ENTRYPOINT ./healthreport -u=${username} -p=${password} -account="/run/secrets/account.json" -email="/run/secrets/email.json"

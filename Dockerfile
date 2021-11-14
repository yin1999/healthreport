FROM golang:alpine as builder

WORKDIR /app

COPY . ./
# static linking
ENV CGO_ENABLED=0
RUN go mod download && go run _script/make.go

FROM busybox:latest

COPY --from=builder /app/healthreport /usr/local/bin/
# add cert file
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

VOLUME ["/run/secrets"]

CMD ["sh", "-c", "healthreport -u=${username} -p=${password} -t=${time} -c=${attempts} -account=/run/secrets/account.json -email=/run/secrets/email.json"]

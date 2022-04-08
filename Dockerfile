FROM golang:latest as builder

RUN apt update && apt install libtesseract-dev -y

WORKDIR /app
COPY . ./

RUN curl https://raw.githubusercontent.com/Shreeshrii/tessdata_shreetest/226419f02431675e24c9937643ce42f3675e2b56/digits.traineddata -o digits.traineddata

RUN go mod download && go run _script/make.go

# images for deployment
FROM debian:stable-slim

RUN apt update && apt install libtesseract4 -y && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/healthreport /usr/local/bin/
COPY --from=builder /app/digits.traineddata /usr/share/tesseract-ocr/4.00/tessdata/

VOLUME ["/run/secrets"]

CMD ["sh", "-c", "healthreport -u=${username} -p=${password} -t=${time} -c=${attempts} -account=/run/secrets/account.json -email=/run/secrets/email.json"]

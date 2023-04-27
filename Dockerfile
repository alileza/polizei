FROM openpolicyagent/conftest:v0.38.0 AS conftest
FROM alpine/helm:3.10.2 AS helm
FROM golang:1.19-alpine AS polizei

COPY . /app

WORKDIR /app

RUN go build -o polizei

# final image
FROM alpine:3.17

COPY --from=conftest /conftest /usr/local/bin/conftest
COPY --from=helm /usr/bin/helm /usr/local/bin/helm
COPY --from=polizei /app/polizei /usr/local/bin/polizei

ENTRYPOINT [ "polizei" ]


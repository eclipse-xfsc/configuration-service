FROM golang:1.24.2-alpine3.20 as build

WORKDIR /app

ADD . .

RUN CGO_ENABLED=0 GOOS=linux go build

FROM alpine:3.15 as runner

COPY --from=build /app/configuration-service /opt/configuration-service

WORKDIR /opt

CMD ["./configuration-service"]

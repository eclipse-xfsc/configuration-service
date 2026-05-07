FROM golang:1.24.2-alpine3.20 as build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o configuration-service .

FROM alpine:3.15

RUN apk --no-cache add ca-certificates

COPY --from=build /app/configuration-service /opt/configuration-service

WORKDIR /opt

EXPOSE 8080

CMD ["./configuration-service"]

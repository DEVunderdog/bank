# Build Stage
FROM golang:1.22.1-alpine3.19 AS builder
WORKDIR /app
COPY  . .
RUN go build -o main main.go
RUN apk add --no-cache curl
RUN version=v4.15.0 && \
    os=linux && \
    arch=amd64 && \
    curl -L https://github.com/golang-migrate/migrate/releases/download/$version/migrate.$os-$arch.tar.gz | tar xvz

# Run stage

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/main .
COPY --from=builder /app/migrate ./migrate
COPY app.env .
COPY start.sh .
COPY wait-for.sh .
COPY database/migration ./migration

EXPOSE 8080
CMD [ "/app/main"]
ENTRYPOINT [ "/app/start.sh" ]
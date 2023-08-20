# Build stage
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o main main.go
RUN apk add curl
RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.16.2/migrate.linux-amd64.tar.gz | tar xvz

# Run stage
FROM alpine
WORKDIR /app
COPY --from=builder /app/main .
COPY --from=builder /app/migrate ./migrate
COPY .env.example .env
COPY db/migration ./migration
COPY start.sh .
COPY wait-for.sh .

EXPOSE 5000
CMD [ "/app/main" ]
ENTRYPOINT [ "/app/start.sh" ]

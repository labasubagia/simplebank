# Build stage
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o main main.go

# Run stage
FROM alpine
WORKDIR /app
COPY --from=builder /app/main .
COPY .env.example .env
COPY db/migration ./db/migration
COPY start.sh .
COPY wait-for.sh .

EXPOSE 5000 5050 6000
CMD [ "/app/main" ]
ENTRYPOINT [ "/app/start.sh" ]

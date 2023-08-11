# Build stage
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o main main.go

# Run stage 
FROM alpine
WORKDIR /app
COPY --from=builder /app/main .

EXPOSE 5000
CMD [ "app/main" ]
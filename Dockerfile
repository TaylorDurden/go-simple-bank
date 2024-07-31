# Build stage
FROM golang:1.22-alpine3.20 as BUILD
WORKDIR /app
COPY . .
RUN go build -o main main.go

# Run stage
FROM alpine:3.13
WORKDIR /app
# only copy BUILD stage /app/main binary file to app directory
COPY --from=BUILD /app/main .

EXPOSE 8080
CMD ["/app/main"]
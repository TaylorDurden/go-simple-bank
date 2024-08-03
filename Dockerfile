# Build stage: This stage is used to build the Go application
FROM golang:1.22-alpine3.19 as BUILD
# Set the working directory in the container to /app
WORKDIR /app
# Copy the current directory (i.e., the application code) into the container
COPY . .
# Build the Go application using the go build command
RUN go build -o main main.go
# Install curl in the container
RUN apk add curl
# Download and extract the migrate tool for database migrations
RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.1/migrate.linux-amd64.tar.gz | tar xvz

# Run stage: This stage is used to create the final Docker image
FROM alpine:3.19
# Set the working directory in the container to /app
WORKDIR /app
# Copy only the main binary file from the BUILD stage into the current stage
COPY --from=BUILD /app/main .
# Copy the migrate tool from the BUILD stage into the current stage
COPY --from=BUILD /app/migrate ./migrate
# Copy the environment variable file into the container
COPY app.env .
# Copy the start script into the container
COPY start.sh .
# Copy the wait-for script into the container
COPY wait-for.sh .
# Copy the database migration files into the container
COPY db/migration ./migration

# Expose port 8080 from the container to the host machine
EXPOSE 8080
# Set the default command to run when the container starts
CMD ["/app/main"]
# Set the entrypoint to the start script, which will be executed when the container starts
ENTRYPOINT [ "/app/start.sh" ]
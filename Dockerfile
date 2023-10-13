# Use the official Go image as a parent image
FROM golang:1.17 AS builder

# Set the working directory to /app
WORKDIR /app

# Copy the current directory contents into the container at /app
COPY . .

# Build the Go application
RUN go build -o go-postgres-api

# Use the official PostgreSQL image as the database
FROM postgres

# Set environment variables for PostgreSQL
ENV POSTGRES_USER asha
ENV POSTGRES_PASSWORD okidoki
ENV POSTGRES_DB mydatabase

# Copy the Go application binary from the builder stage
COPY --from=builder /app/go-postgres-api /app/go-postgres-api

# Expose the API port
EXPOSE 8080

# Run the Go application
CMD ["/app/go-postgres-api"]


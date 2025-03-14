# First stage: Build the Go project
FROM golang:1.22 AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy the Go module files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Build the Go application
RUN CGO_ENABLED=0 GOOS=linux go build -o thdctl .

# Second stage: Run the built executable
FROM alpine:latest

# Install certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Set the working directory inside the container
WORKDIR /app/

# Copy the built executable from the builder stage
COPY --from=builder /app/thdctl .

# Make the binary executable
RUN chmod +x /app/thdctl

# Command to run the executable
CMD ["./thdctl"]

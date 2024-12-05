# Use Go 1.23 or latest stable version
FROM golang:1.23.2-alpine3.19

# Set the working directory in the container
WORKDIR /app

# Install git and other dependencies
RUN apk add --no-cache git

# Copy go.mod and go.sum files first to leverage Docker cache
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Expose the port the app runs on
EXPOSE 8080

# Command to run the application
CMD ["./main"]
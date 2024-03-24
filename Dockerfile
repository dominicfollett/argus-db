# Use an official Golang runtime as the base image
FROM golang:1.21

# Set the working directory in the container
WORKDIR /app

# Copy the Go module files
COPY go.mod ./

# Download the Go module dependencies
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the Go application
RUN go build -o main .

# Expose the port on which the service will run (adjust if necessary)
EXPOSE 8123

# Run the compiled binary when the container starts
CMD ["./main"]
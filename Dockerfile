# Use an official Golang runtime as a parent image
FROM golang:1.22

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Install dependencies
RUN apt-get update && apt-get install -y build-essential

# Build the Go app
RUN go build -o myapp .

# Command to run the executable
CMD ["./myapp"]

# Dockerfile References: https://docs.docker.com/engine/reference/builder/

# Start from the latest golang base image
FROM golang:1.13 as builder

# Add Maintainer Info
LABEL maintainer="vinhut <hutama.alvin@gmail.com>"

# Set Environment Variables
ENV HOME /app
ENV CGO_ENABLED 0
ENV GOOS linux

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app
RUN go build -o main .

FROM golang:1.19.11-alpine

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/* /app/

# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run the executable
CMD ["./main"]

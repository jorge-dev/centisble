# Use build arguments for target OS and architecture
ARG TARGETOS
ARG TARGETARCH

# Build Stage
FROM golang:1.23-alpine AS build

# Declare build arguments in this stage
ARG TARGETOS
ARG TARGETARCH

# Set necessary environment variables for static linking
ENV CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH

# Create and set working directory
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the entire project
COPY . .

# Build the Go application
RUN go build -o main cmd/api/main.go

# Production Stage
FROM gcr.io/distroless/static-debian11 AS prod

# Set working directory
WORKDIR /app

# Copy the statically linked binary and necessary files from the build stage
COPY --from=build /app/main /app/main

# Expose the application port
EXPOSE 8080

# Set a non-root user
USER nonroot:nonroot

# Command to run the application
CMD ["/app/main"]

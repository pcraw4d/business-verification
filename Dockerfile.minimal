# Minimal Dockerfile for Railway deployment
FROM golang:1.22-alpine

WORKDIR /app

# Copy only the main.go file
COPY main.go .

# Build the application
RUN go build -o app main.go

# Expose port
EXPOSE 8080

# Run the application
CMD ["./app"]

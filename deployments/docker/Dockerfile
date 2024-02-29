# First stage: build the application
FROM golang:1.21.4 as builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o mongo-password-rotator ./cmd/rotator

# Second stage: build a small image
FROM alpine:latest

RUN apk --no-cache add ca-certificates

# Create a group and user to run the application
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /home/appuser

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/mongo-password-rotator .

# Change to non-root user
USER appuser

# Command to run the executable
CMD ["./mongo-password-rotator"]

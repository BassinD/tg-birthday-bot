# ==========================================
# Stage 1: Build the Go binary
# ==========================================
FROM golang:1.26-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum first to cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
# (Thanks to //go:embed, prompt.md and the JSON locales will be compiled directly into the binary here)
COPY . .

# Build the binary
# CGO_ENABLED=0 ensures a static binary, which is required for the distroless image
# -trimpath removes file system paths from the compiled executable for cleaner panics/logs
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -o /app/bot ./cmd/bot/main.go

# ==========================================
# Stage 2: Create the final lightweight image
# ==========================================
# We use Google's distroless static image. It contains no shell, no package manager, 
# and no unnecessary tools, making it incredibly secure and tiny.
FROM gcr.io/distroless/static-debian12:nonroot

# Set the working directory
WORKDIR /

# Copy the compiled binary from the builder stage
COPY --from=builder --chown=nonroot:nonroot /app/bot /bot

# Expose the port Cloud Run uses by default
EXPOSE 8080

# Run the container as a non-root user for security
USER nonroot:nonroot

# Set the entrypoint to our binary
ENTRYPOINT ["/bot"]
# Stage 1: Build the Go application
FROM golang:1.27-alpine AS builder

# Install system dependencies needed to compile Go GUI libraries (X11/OpenGL)
RUN apk add --no-cache gcc musl-dev xorg-server-dev libx11-dev libgl1-mesa-dev freeglut-dev

WORKDIR /app

# Copy dependencies first to leverage Docker cache layers
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Compile the standalone Linux binary
RUN CGO_ENABLED=1 GOOS=linux go build -o myguiapp main.go

# Stage 2: Final runtime container
FROM alpine:3.20

# Install runtime dependencies for X11 UI display and netcat (nc) for health checking
RUN apk add --no-cache libx11 libgl1-mesa freeglut mesa-dri-gallium

WORKDIR /app

# Copy the compiled binary and the boot script from the builder stage
COPY --from=builder /app/myguiapp .    
COPY --from=builder /app/boot.go_sh .

# Ensure the script uses standard Unix line endings and is executable
RUN sed -i 's/\r$//' boot.go_sh && chmod +x boot.go_sh

# Trigger the shell script on startup
ENTRYPOINT ["./boot.go_sh"]
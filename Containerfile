# Multi-stage build for photos-ng application
# Stage 1: Build React frontend
FROM docker.io/node:22-alpine AS frontend-builder

ARG GIT_SHA
ENV GIT_SHA=${GIT_SHA}

WORKDIR /app/ui

# Copy source and build
COPY ui/ ./
RUN npm run build

# Stage 2: Build Go backend
FROM docker.io/golang:1.24.6 as backend-builder

ARG GIT_SHA

WORKDIR /app

# Copy go mod files for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-X main.sha=${GIT_SHA}" -a -installsuffix cgo -o photos-ng .

# Stage 3: Final runtime image
FROM docker.io/fedora:41

RUN dnf install -y exiftool

WORKDIR /app

# Copy built application from backend builder
COPY --from=backend-builder /app/photos-ng .

# Copy database migrations
COPY --from=backend-builder /app/internal/datastore/pg/migrations/sql ./migrations/

# Copy built frontend from frontend builder
COPY --from=frontend-builder /app/ui/dist ./ui/dist

# Expose port
EXPOSE 8080

# Default command
CMD ["./photos-ng", "serve", "--server-port=8080", "--server-gin-mode=release", "--server-mode=prod", "--statics-folder=./ui/dist"] 

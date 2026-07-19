# ---- Build stage ----
FROM golang:1.21-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/api-server ./cmd/api

# ---- Runtime stage ----
FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata
COPY --from=build /app/api-server /usr/local/bin/api-server
EXPOSE 8080
ENTRYPOINT ["api-server"]

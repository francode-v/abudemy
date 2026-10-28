# --- build stage ---
FROM golang:1.27-alpine AS build

WORKDIR /src

# Cache dependencies first.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

ENV CGO_ENABLED=0 GOOS=linux
RUN go build -trimpath -ldflags="-s -w" -o /out/abudemy-api ./cmd/api

# --- runtime stage ---
FROM alpine:3.20

RUN adduser -D -u 10001 appuser && apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=build /out/abudemy-api /app/abudemy-api

USER appuser
EXPOSE 8080

ENTRYPOINT ["/app/abudemy-api"]

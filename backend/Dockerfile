FROM golang:1.25-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/api ./cmd/api

FROM alpine:3.20

RUN apk add --no-cache ca-certificates \
	&& adduser -D -H -u 65532 appuser

WORKDIR /app
COPY --from=builder /out/api /app/api

USER appuser

ENV APP_PORT=8080
EXPOSE 8080

ENTRYPOINT ["/app/api"]

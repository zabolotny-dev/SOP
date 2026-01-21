FROM golang:1.25.5-alpine AS builder
WORKDIR /build

COPY hosting-notification-service/go.mod hosting-notification-service/go.sum ./hosting-notification-service/
COPY hosting-contracts/go.mod hosting-contracts/go.sum ./hosting-contracts/
COPY hosting-kit/go.mod hosting-kit/go.sum ./hosting-kit/

WORKDIR /build/hosting-notification-service
RUN go mod download

WORKDIR /build
COPY hosting-contracts/ ./hosting-contracts/
COPY hosting-kit/ ./hosting-kit/
COPY hosting-notification-service/ ./hosting-notification-service/

WORKDIR /build/hosting-notification-service
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /server cmd/server/main.go

FROM alpine:3.21

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

COPY --from=builder /server /server

USER appuser

EXPOSE 8080

CMD ["/server"]
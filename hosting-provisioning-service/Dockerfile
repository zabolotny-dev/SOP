FROM golang:1.24.4-alpine AS builder
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY . .
WORKDIR /app/hosting-provisioning-service
RUN go mod download
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /server cmd/server/main.go

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder --chown=nonroot:nonroot --chmod=755 /server /server

USER nonroot

CMD ["/server"]
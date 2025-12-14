FROM golang:1.24.4-alpine AS builder
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app

COPY . .

WORKDIR /app/hosting-service

RUN go mod download

RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /server cmd/server/main.go
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /migrator cmd/migrator/main.go

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder --chown=nonroot:nonroot --chmod=755 /server /server
COPY --from=builder --chown=nonroot:nonroot --chmod=755 /migrator /migrator

USER nonroot
EXPOSE 8080

CMD ["/server"]
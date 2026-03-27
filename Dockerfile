# ---------- BUILD ----------
FROM golang:1.25-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o /app/bin/app ./cmd/app


# ---------- RUNTIME ----------
FROM alpine:3.20

WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY . /app
COPY --from=builder /app/bin/app /app/bin/app

EXPOSE 8080

ENTRYPOINT ["/app/bin/app"]

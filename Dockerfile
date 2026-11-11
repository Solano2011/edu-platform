FROM golang:1.27-alpine AS builder

ENV GOTOOLCHAIN=auto
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/edu-platform ./cmd/app

FROM alpine:3.20

WORKDIR /app
RUN apk --no-cache add ca-certificates tzdata

COPY --from=builder /app/bin/edu-platform /app/edu-platform
COPY web /app/web
COPY migrations /app/migrations

EXPOSE 8080

CMD ["/app/edu-platform"]
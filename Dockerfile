FROM golang:1.24-alpine AS builder

WORKDIR /src

# cache deps
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build static binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-s -w" -o /app/egolang ./main.go

FROM alpine:3.21
RUN apk add --no-cache ca-certificates
COPY --from=builder /app/egolang /usr/local/bin/egolang

EXPOSE 9000
ENV APP_PORT=9000
CMD ["/usr/local/bin/egolang"]

FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o hederium ./cmd/server

FROM alpine:3.17
WORKDIR /app
COPY --from=builder /app/hederium .
COPY --from=builder /app/configs ./configs
EXPOSE 8080
ENTRYPOINT ["./hederium"]

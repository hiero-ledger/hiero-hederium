FROM golang:1.22-alpine@sha256:1699c10032ca2582ec89a24a1312d986a3f094aed3d5c1147b19880afe40e052 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o hederium ./cmd/server

FROM alpine:3.17@sha256:8fc3dacfb6d69da8d44e42390de777e48577085db99aa4e4af35f483eb08b989
WORKDIR /app
COPY --from=builder /app/hederium .
COPY --from=builder /app/configs ./configs
EXPOSE 7546
ENTRYPOINT ["./hederium"]

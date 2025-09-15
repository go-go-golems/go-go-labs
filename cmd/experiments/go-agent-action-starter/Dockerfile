# Build stage
FROM golang:1.22-alpine AS builder
WORKDIR /src
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -o /out/agent-action ./cmd/agent-action

# Run stage
FROM alpine:3.20
RUN apk add --no-cache ca-certificates bash
WORKDIR /home/app
COPY --from=builder /out/agent-action /usr/local/bin/agent-action
ENTRYPOINT ["/usr/local/bin/agent-action"]

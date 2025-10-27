# docker/redirect-server/Dockerfile
FROM golang:1.19-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o redirect-server ./cmd/redirect-server

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

COPY --from=builder /app/redirect-server .
COPY --from=builder /app/configs ./configs

RUN mkdir -p logs

EXPOSE 8080

CMD ["./redirect-server"]
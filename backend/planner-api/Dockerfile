FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY backend/planner-api/go.mod backend/planner-api/go.sum ./
RUN go mod download

COPY backend/planner-api ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /planner-api ./cmd/server

FROM alpine:3.22

RUN adduser -D -g '' appuser

USER appuser
WORKDIR /app

COPY --from=builder /planner-api /planner-api

EXPOSE 8000

ENTRYPOINT ["/planner-api"]

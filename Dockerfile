FROM golang:1.24.3 AS builder

WORKDIR /app

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o todo-app .

# --- Final stage ---
FROM alpine:3.20

WORKDIR /app

ENV TODO_PORT=7540
ENV TODO_DBFILE=/data/scheduler.db
ENV TODO_PASSWORD=12345

RUN mkdir -p /data

COPY --from=builder /app/todo-app /app/
COPY web ./web

EXPOSE 7540

CMD ["/app/todo-app"]

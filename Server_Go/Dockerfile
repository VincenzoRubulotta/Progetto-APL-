FROM golang:1.25.4 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o scopone_server .

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/scopone_server .

EXPOSE 50051

CMD ["./scopone_server"]
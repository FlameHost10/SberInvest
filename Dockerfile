FROM golang:1.23 AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o myapp .

RUN ls -l /app

FROM alpine:latest

RUN apk --no-cache add ca-certificates

COPY --from=builder /app/myapp /myapp

RUN chmod +x /myapp

WORKDIR /

CMD ["./myapp"]

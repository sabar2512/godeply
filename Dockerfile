# build stage
FROM golang:1.25 AS builder
WORKDIR /app

# copy go modules
COPY go.mod go.sum ./
RUN go mod download

# copy semua file project
COPY . .

# build binary untuk linux
RUN go build -o main .

# run stage
FROM debian:bookworm

WORKDIR /root/

COPY --from=builder /app/main .

EXPOSE 8080

CMD ["./main"]

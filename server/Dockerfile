FROM golang:1.20 AS prep

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /deployer-server

FROM alpine:latest

RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=prep /deployer-server .

EXPOSE 8080
CMD ["./deployer-server"]
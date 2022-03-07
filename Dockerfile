# build stage
FROM golang:1.17.7-alpine3.14 AS builder
WORKDIR /app
COPY . .
RUN go build -o main main.go

# run stage
FROM alpine:3.14 as deploy
WORKDIR /app
COPY --from=builder /app/main .
COPY app.env .

ENTRYPOINT [ "/app/main" ]

FROM golang:alpine AS builder

WORKDIR /app   

# COPY go.mod go.sum ./
# RUN go mod tidy && go mod download

COPY go.mod go.sum main.go ./

RUN go build -o main .

FROM alpine:latest

COPY --from=builder /app/main .

CMD [ "./main" ]
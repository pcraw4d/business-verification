FROM golang:1.22-alpine

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o kyb-business-intelligence-gateway main.go

EXPOSE 8087

CMD ["./kyb-business-intelligence-gateway"]

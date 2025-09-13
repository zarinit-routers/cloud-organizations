FROM golang:latest

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o srv cmd/organizations/main.go 

EXPOSE 8060

CMD ["/app/srv"]

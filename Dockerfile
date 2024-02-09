FROM golang:1.21.1

WORKDIR /forum/

COPY . .

RUN GOOS=linux GOARCH=amd64 go build -o main

CMD ["./main"]
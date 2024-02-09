FROM golang:1.21.1

WORKDIR /forum/

RUN apt-get update && apt-get install -y \
    sqlite3 \
    libsqlite3-dev

COPY . .

ENV CGO_ENABLED=1

RUN go build -o main

CMD ["./main"]
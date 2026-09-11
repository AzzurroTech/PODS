FROM golang:1.20-alpine

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .

RUN go build -o song-server ./main.go

EXPOSE 8083

CMD ["/app/song-server", "--port=8083"]

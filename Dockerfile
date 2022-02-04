FROM golang:1.17.0-bullseye

RUN go version

ENV GOPATH=/

COPY ./ ./

RUN go mod download

RUN go build -o tg-connection-base ./main.go

CMD ["./tg-connection-base"]
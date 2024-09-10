FROM golang:1.23-bookworm

WORKDIR /go/src/bot

CMD [ "go", "run", "cmd/server/main.go" ]

FROM golang:1.27-bookworm

WORKDIR /go/src/bot

CMD [ "go", "run", "cmd/bot/main.go" ]

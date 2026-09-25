FROM golang:1.27-bookworm

WORKDIR /go/src/bot

CMD [ "go", "run", "cmd/server/main.go" ]

FROM golang:1.24-bookworm

WORKDIR /go/src/bot

CMD [ "go", "run", "cmd/server/main.go" ]

FROM golang:1.22-bookworm

WORKDIR /go/src/bot

CMD [ "go", "run", "main.go" ]

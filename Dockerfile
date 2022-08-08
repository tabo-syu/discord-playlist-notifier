FROM golang:1.18-bullseye

WORKDIR /go/src/bot

CMD [ "go", "run", "main.go" ]

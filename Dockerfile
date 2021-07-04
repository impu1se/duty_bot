FROM golang:1.13

RUN mkdir /duty_bot
ADD . /duty_bot/
WORKDIR /duty_bot

RUN go mod download
RUN go build -o duty_bot cmd/duty_bot/main.go

CMD ["/duty_bot/duty_bot"]

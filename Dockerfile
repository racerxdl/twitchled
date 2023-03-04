FROM golang:1.18-alpine as build

RUN apk update

RUN apk add git ca-certificates

ADD . /go/src/github.com/racerxdl/twitchled

ENV GO111MODULE=on


WORKDIR /go/src/github.com/racerxdl/twitchled/cmd
RUN go get -v
RUN GOOS=linux go build -o ../twitchled

FROM alpine:latest

RUN apk --no-cache add ca-certificates

RUN mkdir -p /opt/
WORKDIR /opt/

COPY --from=build /go/src/github.com/racerxdl/twitchled/twitchled .

CMD /opt/twitchled
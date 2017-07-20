FROM golang:1.8-alpine3.6 as builder

ADD . /go/src/github.com/dominikschulz/vanity
WORKDIR /go/src/github.com/dominikschulz/vanity

RUN go install

FROM alpine:3.6

COPY --from=builder /go/bin/vanity /usr/local/bin/vanity
CMD [ "/usr/local/bin/vanity" ]
EXPOSE 8080 8081

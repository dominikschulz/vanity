FROM golang:1.5

ENV GOBIN /go/bin
ENV GOPATH /go
ADD . /go/src/github.com/dominikschulz/vanity
WORKDIR /go/src/github.com/dominikschulz/vanity
RUN go get ./...
RUN go install

CMD [ "/go/bin/vanity" ]

ENV VANITY_LISTEN ":8080"

EXPOSE 8080

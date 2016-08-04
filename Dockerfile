FROM alpine:latest

ADD vanity /usr/local/bin/vanity
CMD [ "/usr/local/bin/vanity" ]
EXPOSE 8080 8081

FROM alpine:latest

WORKDIR /

COPY /tmp/admission-webhook /admission-webhook

ENTRYPOINT ["./admission-webhook"]

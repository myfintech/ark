FROM alpine:latest

RUN apk update && apk add --no-cache ca-certificates && update-ca-certificates

ARG GIT_SHA
ENV GIT_SHA=${GIT_SHA}

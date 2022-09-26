FROM golang:1.14-alpine as deps
WORKDIR /opt/app
RUN apk update
RUN apk upgrade
RUN apk add bash
RUN apk add --no-cache git
COPY ./go.mod ./go.mod
COPY ./go.sum ./go.sum
RUN go mod download

FROM deps as build
ENV CGO_ENABLED=0
ENV GO111MODULE=on
ENV GOFLAGS -mod=vendor
WORKDIR /opt/app
COPY ./src/go/lib ./src/go/lib
COPY ./src/go/tools/arkcli ./src/go/tools/arkcli
RUN go mod vendor
RUN go build -o ark-linux ./src/go/tools/arkcli/main.go
RUN GOOS=darwin go build -o ark-macos ./src/go/tools/arkcli/main.go

FROM alpine:latest as release
WORKDIR /usr/bin
COPY --from=build /opt/app/ark-linux /usr/bin/ark
RUN chmod +x /usr/bin/ark
ENTRYPOINT ["/usr/bin/ark", "build", "--watch"]

FROM scratch as output
COPY --from=build /opt/app/ark-linux .
COPY --from=build /opt/app/ark-macos .

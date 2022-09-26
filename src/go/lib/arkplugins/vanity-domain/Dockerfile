# syntax = docker/dockerfile:experimental
FROM gcr.io/managed-infrastructure/mantl/sre/go-modules:cf198416b2b3b1e83afa92aa6bee58cacea6a61c2f95b995a7a8ff56a2e1117c as modules
FROM gcr.io/managed-infrastructure/mantl/sre/golang.1.17-alpine:e06ba6501578d1609935a27404710102a35461f5a85b3bb1f32cc27b71bc7c0e as build

WORKDIR /opt/app
COPY --from=modules /go/pkg/mod /go/pkg/mod
COPY --from=modules /opt/app /opt/app

COPY ./ ./

ENV CGO_ENABLED=0
RUN go mod vendor

RUN --mount=type=cache,target=/root/.cache \
    go build -o vanity-domain-plugin /opt/app/src/go/lib/arkplugins/vanity-domain/main.go

FROM scratch as output
COPY --from=build /opt/app/vanity-domain-plugin /usr/local/bin/
ENTRYPOINT ["/usr/local/bin/vanity-domain-plugin", "-"]

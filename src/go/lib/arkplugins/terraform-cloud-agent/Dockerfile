# syntax = docker/dockerfile:experimental
FROM ${modules} as modules
FROM ${from} as build
WORKDIR /opt/app
COPY --from=modules /go/pkg/mod /go/pkg/mod
COPY --from=modules /opt/app /opt/app

COPY ./ ./

ENV CGO_ENABLED=0
RUN go mod vendor

# Compile binary for linux
RUN --mount=type=cache,target=/root/.cache \
  go build -o ${package.name} /opt/app/src/go/lib/arkplugins/terraform-cloud-agent/*.go

FROM scratch as output
COPY --from=build /opt/app/${package.name} /usr/local/bin/
ENTRYPOINT ["/usr/local/bin/${package.name}"]

# syntax=docker/dockerfile:1.0-experimental
FROM ${from} as build

%{ for file in sources ~}
COPY ${file.src} ${file.dest}
%{ endfor ~}

%{ for cmd in commands ~}
RUN ${cmd}
%{ endfor ~}

FROM scratch
COPY --from=build /Dockerfile /Dockerfile
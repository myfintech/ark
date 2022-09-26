FROM alpine:3.9
RUN apk --no-cache add bash jq curl
COPY --from=vault /bin/vault /usr/local/bin/vault
WORKDIR /opt/vault

COPY ./src/docker/vault ./

ENV VAULT_DATA_VOLUME /mnt/vault
ENV VAULT_BASE_STORAGE /opt/vault/data
ENV VAULT_ADDR http://localhost:8200
ENV VAULT_CONFIG /opt/vault/config.hcl

CMD ["/opt/vault/entrypoint.sh"]
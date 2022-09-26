FROM redis:4.0.14

COPY ./redis.conf /usr/local/etc/redis.conf
COPY ./entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]

CMD ["redis-server", "/usr/local/etc/redis.conf"]

FROM stackato/base-redis

ADD . /source
ADD stackon.json /

CMD ["redis-server", "/source/redis-server.conf"]

EXPOSE 6464
VOLUME ["/var/stackato/data/applog_redis"]

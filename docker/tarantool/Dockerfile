FROM tarantool/tarantool:2.10.0

WORKDIR /opt/tarantool

COPY docker/tarantool/init.lua /opt/tarantool/init.lua

RUN mkdir -p /var/lib/tarantool

RUN chown -R tarantool:tarantool /var/lib/tarantool /opt/tarantool

USER tarantool

EXPOSE 3301

CMD ["tarantool", "/opt/tarantool/init.lua"]
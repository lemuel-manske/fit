FROM rabbitmq:4-management

COPY docker/rabbitmq.sh /usr/local/bin/fit-rabbitmq.sh

HEALTHCHECK --interval=5s --timeout=5s --start-period=10s --retries=12 \
    CMD test -f /tmp/fit-rabbitmq-ready && rabbitmq-diagnostics -q check_running


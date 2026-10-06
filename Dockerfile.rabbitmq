FROM rabbitmq:4-management

HEALTHCHECK --interval=5s --timeout=5s --start-period=10s --retries=12 \
    CMD rabbitmq-diagnostics -q check_running

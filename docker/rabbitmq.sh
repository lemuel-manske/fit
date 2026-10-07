#!/bin/sh
set -eu

docker-entrypoint.sh rabbitmq-server &
broker_pid=$!
trap 'kill -TERM "$broker_pid"; wait "$broker_pid"' TERM INT

rm -f /tmp/fit-rabbitmq-ready
rabbitmqctl await_startup
# The demonstration user is restricted to Fit resources in the dedicated vhost.
rabbitmqctl set_user_tags "$RABBITMQ_DEFAULT_USER"
rabbitmqctl set_permissions -p fit "$RABBITMQ_DEFAULT_USER" \
  '^(fit\..*|amq\.gen-.*)$' '^(fit\..*|amq\.gen-.*|amq\.default)$' '^(fit\..*|amq\.gen-.*)$'
touch /tmp/fit-rabbitmq-ready
wait "$broker_pid"

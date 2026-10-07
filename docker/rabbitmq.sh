#!/bin/sh
set -eu

docker-entrypoint.sh rabbitmq-server &
broker_pid=$!
trap 'kill -TERM "$broker_pid"; wait "$broker_pid"' TERM INT

rm -f /tmp/fit-rabbitmq-ready
attempts=0
until rabbitmqctl await_startup >/dev/null 2>&1; do
  attempts=$((attempts + 1))
  if [ "$attempts" -ge 60 ] || ! kill -0 "$broker_pid" 2>/dev/null; then
    exit 1
  fi
  sleep 1
done
# The demonstration user is restricted to Fit resources in the dedicated vhost.
rabbitmqctl set_user_tags "$RABBITMQ_DEFAULT_USER"
rabbitmqctl set_permissions -p fit "$RABBITMQ_DEFAULT_USER" \
  '^(fit\..*|amq\.gen-.*)$' '^(fit\..*|amq\.gen-.*|amq\.default)$' '^(fit\..*|amq\.gen-.*)$'
touch /tmp/fit-rabbitmq-ready
wait "$broker_pid"

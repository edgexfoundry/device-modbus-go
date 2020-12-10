#!/bin/sh -e

EDGEX_STARTUP_DURATION=$(snapctl get startup-duration)

if [ -n "$EDGEX_STARTUP_DURATION" ]; then
  export EDGEX_STARTUP_DURATION
fi

EDGEX_STARTUP_INTERVAL=$(snapctl get startup-interval)

if [ -n "$EDGEX_STARTUP_INTERVAL" ]; then
  export EDGEX_STARTUP_INTERVAL
fi

exec "$@"

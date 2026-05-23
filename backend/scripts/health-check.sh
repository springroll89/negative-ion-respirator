#!/bin/bash

HEALTH_URL="${HEALTH_URL:-https://localhost}"
TIMEOUT=5
COMPOSE_FILE="/opt/ion-respirator/backend/docker-compose.prod.yml"

# Check HTTPS endpoint
if curl -sk --max-time "$TIMEOUT" "$HEALTH_URL" > /dev/null 2>&1; then
    echo "OK: Health check passed"
else
    echo "CRITICAL: Health check failed - $HEALTH_URL unreachable"
    exit 1
fi

# Check Docker services
if [ -f "$COMPOSE_FILE" ]; then
    RUNNING=$(docker compose -f "$COMPOSE_FILE" ps --status running -q 2>/dev/null | wc -l)
    EXPECTED=6
    if [ "$RUNNING" -lt "$EXPECTED" ]; then
        echo "CRITICAL: Only $RUNNING/$EXPECTED services running"
        exit 1
    fi
    echo "OK: All $RUNNING services running"
else
    echo "WARNING: Compose file not found at $COMPOSE_FILE"
fi

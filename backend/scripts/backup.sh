#!/bin/bash
set -e

BACKUP_DIR="/opt/backups/ion-respirator"
RETENTION_DAYS=30

mkdir -p "$BACKUP_DIR"
DATE=$(date +%Y%m%d_%H%M%S)
FILE="$BACKUP_DIR/ion_backup_$DATE.sql.gz"

cd /opt/ion-respirator/backend
docker compose exec -T postgres pg_dump -U ion ion_respirator | gzip > "$FILE"

echo "Backup saved: $FILE ($(du -h "$FILE" | cut -f1))"

# Clean old backups
find "$BACKUP_DIR" -name "*.sql.gz" -mtime +$RETENTION_DAYS -delete
echo "Cleaned backups older than $RETENTION_DAYS days"

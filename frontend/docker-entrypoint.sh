#!/bin/sh
set -e

cat > /usr/share/nginx/html/config.json <<EOF
{
  "apiUrl": "${VITE_API_URL}",
  "wsUrl": "${VITE_WS_URL}"
}
EOF

exec "$@"

#!/bin/bash
# Syspulse dev start -- builds frontend, then runs backend with Air hot-reload
export PATH=$PATH:/usr/local/go/bin:/root/go/bin
cd /opt/syspulse

pkill -f air 2>/dev/null
pkill -f vite 2>/dev/null

echo "==> Building frontend and copying to internal/api/..."
make frontend 2>&1

echo ""
echo "==> Starting backend with Air hot-reload on port 7655..."
air &
BACK_PID=$!

echo ""
echo "PROX-WEB Syspulse:"
echo "  Backend: http://192.168.0.87:7655"
echo "  PID: $BACK_PID"
echo "  Stop: pkill -f air"
echo "  Logs: tail -f /opt/syspulse/tmp/build-errors.log"

wait

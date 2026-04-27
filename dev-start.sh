#!/bin/bash
# Syspulse dev start — runs backend (Air hot-reload) + frontend (Vite) concurrently
export PATH=$PATH:/usr/local/go/bin:/root/go/bin
cd /opt/syspulse

# Kill any leftover processes
pkill -f "air" 2>/dev/null
pkill -f "vite" 2>/dev/null

echo "Starting frontend (Vite) on port 5173..."
cd frontend-modern && npm run dev -- --host 0.0.0.0 &
FRONT_PID=$!
cd ..

echo "Starting backend (Air hot-reload) on port 7655..."
air &
BACK_PID=$!

echo ""
echo "PROX-WEB Syspulse dev environment:"
echo "  Frontend (Vite):  http://192.168.0.87:5173"
echo "  Backend (API):    http://192.168.0.87:7655"
echo ""
echo "PIDs: frontend=$FRONT_PID backend=$BACK_PID"
echo "Stop: pkill -f air; pkill -f vite"

wait

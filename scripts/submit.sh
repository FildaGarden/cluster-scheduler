#!/bin/bash

MASTER="http://localhost:8080"
COMMAND="echo 'Testovaci uloha spoustena pres script'; sleep 5; echo 'Uloha hotova'"
CPU=1
MEM=128
RATE=5  # req/sec
INTERVAL=$(echo "scale=4; 1/$RATE" | bc)

echo "🚀 Odesílám $RATE req/sec na Mastera ($MASTER)..."

while true; do
    curl -s -X POST "$MASTER/submit" -d "{
      \"Command\": \"$COMMAND\",
      \"CPUCores\": $CPU,
      \"MemoryMB\": $MEM
    }" &
    sleep "$INTERVAL"
done

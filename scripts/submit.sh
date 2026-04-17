#!/bin/bash

# Adresa tvého Mastera (změň, pokud se změní IP)
MASTER="http://localhost:8080"

# Statické parametry pro rychlé testování
COMMAND="echo 'Testovaci uloha spoustena pres script'; sleep 5; echo 'Uloha hotova'"
CPU=1
MEM=128

echo "🚀 Odesílám testovací úlohu na Mastera ($MASTER)..."

curl -X POST "$MASTER/submit" -d "{
  \"Command\": \"$COMMAND\",
  \"CPUCores\": $CPU,
  \"MemoryMB\": $MEM
}"

echo -e "\n✅ Odesláno."

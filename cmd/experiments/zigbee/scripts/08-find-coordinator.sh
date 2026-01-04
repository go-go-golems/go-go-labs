#!/usr/bin/env bash
set -euo pipefail

echo "=== lsusb (filtered) ==="
lsusb | rg -i 'zigbee|sonoff|itead|10c4:ea60|cp210|silicon labs|cc2652|texas|ti' || true

echo
echo "=== /dev serial candidates ==="
if [ -d /dev/serial/by-id ]; then
  ls -l /dev/serial/by-id || true
else
  echo "/dev/serial/by-id does not exist"
fi

echo
echo "=== tty devices ==="
ls -l /dev/ttyUSB* /dev/ttyACM* 2>/dev/null || echo "No /dev/ttyUSB* or /dev/ttyACM* devices found"


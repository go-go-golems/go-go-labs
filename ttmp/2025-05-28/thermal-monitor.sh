#!/bin/bash

# Framework Laptop Thermal Monitoring Script
# Created: 2025-05-28
# Purpose: Quick thermal status check for Manuel's Framework laptop

echo "=== Framework Laptop Thermal Status ==="
echo "Date: $(date)"
echo

# Check fw-fanctrl service status
echo "🔧 fw-fanctrl Service Status:"
if systemctl is-active --quiet fw-fanctrl; then
    echo "  ✅ Service is running"
    echo "  Strategy: $(fw-fanctrl print current | grep "Strategy in use" | awk '{print $4}' | tr -d "'")"
    echo "  Fan Speed: $(fw-fanctrl print speed | grep "Current fan speed" | awk '{print $4}' | tr -d "'")"
else
    echo "  ❌ Service is NOT running"
fi
echo

# Check temperatures
echo "🌡️  Current Temperatures:"
sensors | grep -A 5 coretemp | grep -E "(Package|Core)" | while read line; do
    temp=$(echo "$line" | awk '{print $3}' | tr -d '+°C')
    if (( $(echo "$temp > 80" | bc -l) )); then
        echo "  🔥 $line"
    elif (( $(echo "$temp > 70" | bc -l) )); then
        echo "  ⚠️  $line"
    else
        echo "  ✅ $line"
    fi
done
echo

# Check fan RPM
echo "🌪️  Fan Status:"
if command -v ectool >/dev/null 2>&1; then
    fan_rpm=$(sudo ectool pwmgetfanrpm 2>/dev/null | awk '{print $4}')
    if [ ! -z "$fan_rpm" ]; then
        echo "  Fan RPM: $fan_rpm"
        if [ "$fan_rpm" -gt 7000 ]; then
            echo "  Status: 🔥 High speed (cooling)"
        elif [ "$fan_rpm" -gt 4000 ]; then
            echo "  Status: ⚠️  Medium speed"
        else
            echo "  Status: ✅ Low speed (quiet)"
        fi
    else
        echo "  ❌ Cannot read fan RPM"
    fi
else
    echo "  ❌ ectool not available"
fi
echo

# Check for recent crashes
echo "💥 System Stability:"
last_reboot=$(last reboot | head -1 | awk '{print $5, $6, $7, $8}')
echo "  Last reboot: $last_reboot"

# Check for thermal throttling
thermal_events=$(dmesg | grep -i thermal | wc -l)
if [ "$thermal_events" -gt 0 ]; then
    echo "  ⚠️  $thermal_events thermal events in dmesg"
else
    echo "  ✅ No thermal events in dmesg"
fi
echo

# Check EC communication errors (last hour)
echo "🔌 EC Communication:"
ec_errors=$(sudo journalctl -u fw-fanctrl --since "1 hour ago" 2>/dev/null | grep -c "Bad message\|INVALID_CHECKSUM" || echo "0")
if [ "$ec_errors" -gt 10 ]; then
    echo "  ⚠️  $ec_errors EC errors in last hour (high)"
elif [ "$ec_errors" -gt 0 ]; then
    echo "  ℹ️  $ec_errors EC errors in last hour (normal)"
else
    echo "  ✅ No EC errors in last hour"
fi
echo

# Recommendations
echo "💡 Quick Actions:"
echo "  Check detailed logs: sudo journalctl -u fw-fanctrl -f"
echo "  Switch strategies: fw-fanctrl use [lazy|medium|agile|deaf]"
echo "  Manual fan test: sudo ectool fanduty 50"
echo "  Stress test: stress --cpu 4 --timeout 30s"
echo

echo "=== End of Report ===" 
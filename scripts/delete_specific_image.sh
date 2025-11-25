#!/bin/bash
# Delete a specific BCM software image by UUID
# Usage: ./delete_specific_image.sh <uuid>
# Example: ./delete_specific_image.sh 128dbc2d-fa1c-4e9f-bbae-9e47ec400c6c

set -e

# Configuration
BCM_ENDPOINT="${BCM_ENDPOINT:-https://172.21.15.254:8081}"
BCM_USERNAME="${BCM_USERNAME:-root}"
BCM_PASSWORD="${BCM_PASSWORD:-Hashicorp123!}"

# Target UUID
TARGET_UUID="${1:-128dbc2d-fa1c-4e9f-bbae-9e47ec400c6c}"

echo "🗑️  Deleting BCM Software Image"
echo "================================"
echo "UUID: $TARGET_UUID"
echo "Endpoint: $BCM_ENDPOINT"
echo ""

# Login to BCM
echo "🔐 Authenticating with BCM..."
LOGIN_RESPONSE=$(curl -s -k -c /tmp/bcm_cookies.txt -X POST "$BCM_ENDPOINT/json" \
  -H "Content-Type: application/json" \
  -d "{\"service\":\"login\",\"username\":\"$BCM_USERNAME\",\"password\":\"$BCM_PASSWORD\"}")

if [ "$LOGIN_RESPONSE" != "true" ]; then
  echo "❌ Login failed: $LOGIN_RESPONSE"
  exit 1
fi

echo "✅ Authenticated successfully"
echo ""

# Remove the software image
echo "🗑️  Deleting software image..."
DELETE_RESPONSE=$(curl -s -k -b /tmp/bcm_cookies.txt -X POST "$BCM_ENDPOINT/json" \
  -H "Content-Type: application/json" \
  -d "{\"service\":\"CMPart\",\"call\":\"removeSoftwareImage\",\"args\":[\"$TARGET_UUID\",false,false,false]}")

echo "Response: $DELETE_RESPONSE"
echo ""

# Check if deletion was successful
if echo "$DELETE_RESPONSE" | grep -q "\"success\":false"; then
  echo "❌ Deletion failed"
  echo "$DELETE_RESPONSE" | python3 -m json.tool 2>/dev/null || echo "$DELETE_RESPONSE"
  rm -f /tmp/bcm_cookies.txt
  exit 1
fi

if [ "$DELETE_RESPONSE" == "true" ] || echo "$DELETE_RESPONSE" | grep -q "\"success\":true"; then
  echo "✅ Software image deleted successfully"
else
  echo "⚠️  Unexpected response (may still be successful):"
  echo "$DELETE_RESPONSE" | python3 -m json.tool 2>/dev/null || echo "$DELETE_RESPONSE"
fi

# Cleanup
rm -f /tmp/bcm_cookies.txt

echo ""
echo "✅ Done!"

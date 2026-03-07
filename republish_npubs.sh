#!/bin/bash
# Republish profiles and recent notes for whitelisted npubs to all blastr relays

NPUBS=(
  "npub10z4f5xu9g4rjjvqx4zhhn9hnrqddpczlchq4w3v0wn6430u838hqdlx7l3"
  "npub177fz5zkm87jdmf0we2nz7mm7uc2e7l64uzqrv6rvdrsg8qkrg7yqx0aaq7"
)

# Build relay list from relays_blastr.json
RELAYS=$(jq -r '.[]' /home/server/haven/relays_blastr.json | sed 's|^|wss://|' | tr '\n' ' ')

# Source relays to fetch from
SOURCE_RELAYS="wss://relay.damus.io wss://relay.primal.net wss://relay.nostr.band wss://nos.lol wss://podtards.com"

for NPUB in "${NPUBS[@]}"; do
  echo "=== Processing $NPUB ==="

  # Decode npub to hex
  HEX=$(nak decode "$NPUB" 2>/dev/null)
  echo "Hex pubkey: $HEX"

  # Fetch profile (kind 0) and republish
  echo "Fetching profile (kind 0)..."
  PROFILE=$(nak req -k 0 -a "$HEX" -l 1 $SOURCE_RELAYS 2>/dev/null | head -1)
  if [ -n "$PROFILE" ]; then
    echo "Publishing profile to blastr relays..."
    echo "$PROFILE" | nak event $RELAYS 2>&1 | tail -5
    echo "Profile published."
  else
    echo "No profile found."
  fi

  # Fetch recent notes (kind 1) and republish
  echo "Fetching recent notes (kind 1, last 200)..."
  EVENTS=$(nak req -k 1 -a "$HEX" -l 200 $SOURCE_RELAYS 2>/dev/null | sort -u)
  COUNT=$(echo "$EVENTS" | grep -c '^{')
  echo "Found $COUNT unique events."

  if [ "$COUNT" -gt 0 ]; then
    echo "Publishing events to blastr relays..."
    echo "$EVENTS" | while read -r event; do
      [ -z "$event" ] && continue
      echo "$event" | nak event $RELAYS 2>/dev/null
    done
    echo "Events published."
  fi

  # Fetch contact list (kind 3) and republish
  echo "Fetching contact list (kind 3)..."
  CONTACTS=$(nak req -k 3 -a "$HEX" -l 1 $SOURCE_RELAYS 2>/dev/null | head -1)
  if [ -n "$CONTACTS" ]; then
    echo "Publishing contact list to blastr relays..."
    echo "$CONTACTS" | nak event $RELAYS 2>&1 | tail -5
    echo "Contact list published."
  else
    echo "No contact list found."
  fi

  # Fetch relay list (kind 10002) and republish
  echo "Fetching relay list (kind 10002)..."
  RELAYLIST=$(nak req -k 10002 -a "$HEX" -l 1 $SOURCE_RELAYS 2>/dev/null | head -1)
  if [ -n "$RELAYLIST" ]; then
    echo "Publishing relay list to blastr relays..."
    echo "$RELAYLIST" | nak event $RELAYS 2>&1 | tail -5
    echo "Relay list published."
  else
    echo "No relay list found."
  fi

  echo ""
done

echo "=== Done ==="

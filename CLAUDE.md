# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Run

```bash
go build                     # Build binary
./haven                      # Start relay server (default port 3355)
./haven backup               # Backup all DBs to zip
./haven backup -r outbox     # Export single relay to JSONL
./haven backup --to-cloud    # Backup and upload to S3
./haven restore              # Restore from zip
./haven restore --from-cloud # Download and restore from S3
./haven import               # Import notes from seed relays
```

No test suite exists. No Makefile — use `go build` directly or goreleaser for cross-platform releases.

## Architecture

Haven runs four Nostr relays on a single HTTP server, each at a different path, plus an integrated Blossom media server:

| Relay | Path | Auth Required | Who Can Write | Content |
|-------|------|---------------|---------------|---------|
| Private | /private | Yes | Whitelisted only | Any kind (drafts, eCash) |
| Chat | /chat | Yes | WoT members | DMs, group chat, gift wraps |
| Inbox | / (default) | No (read) | WoT, must tag owner | Reactions, zaps, replies |
| Outbox | / (default) | No (read) | Owner/whitelisted | Public notes, media |

Each relay has its own separate database instance (5 total: private, chat, outbox, inbox, blossom).

### Request Flow

`main.go` routes HTTP requests → khatru relay framework handles WebSocket/NIP protocol → `policies.go` enforces auth + access control → eventstore persists to LMDB/BadgerDB → `blastr.go` publishes outbox events to external relays.

### Key Files

- **main.go** — Entry point, HTTP routing, CLI command dispatch, goroutine orchestration
- **config.go** — All config loading from `.env` and JSON files, Config struct definition
- **init.go** — Relay + database initialization, policy attachment, Blossom setup, HTML rendering
- **policies.go** — Access control: whitelist checks, WoT enforcement, blacklist, content validation
- **backup.go** — Zip/JSONL export/import, cloud upload/download, periodic backup scheduler
- **import.go** — Fetches owner notes and tagged notes from seed relays in batched time windows
- **jsonl.go** — JSONL serialization with deduplication and large buffer support (100MB)
- **blastr.go** — Async event publishing to configured external relays
- **limits.go** — Per-relay rate limiting config (event IP, connection rate, filter complexity)
- **pkg/wot/** — Web of Trust: levels 0-3, in-memory pubkey graph with atomic refresh
- **internal/cloud/** — S3-compatible cloud provider abstraction (upload/download interfaces)
- **ddns.go** — Periodic DDNS updater: detects public IP changes and updates DNS via provider API
- **internal/ddns/** — DDNS provider interface + implementations (Cloudflare, DuckDNS, No-IP)

### Database

Two backends available via `DB_ENGINE` env var:
- **LMDB** (default) — Better performance, recommended for NVMe
- **BadgerDB** — Alternative embedded store

Each relay type gets a separate database directory under `db/`.

### Web of Trust (WoT)

Configured via `WOT_DEPTH` (0=disabled, 1=owner-only, 2=direct follows, 3=follows-of-follows with minimum follower threshold). Controls who can write to chat and inbox relays. Refreshes periodically from seed relays. Implementation in `pkg/wot/simple_in_memory.go` uses atomic maps for thread-safe O(1) lookups.

### Access Control

- **Whitelist** (`whitelisted_npubs.json`) — Full access to all relays + Blossom uploads. Owner npub always included.
- **Blacklist** (`blacklisted_npubs.json`) — Completely blocked from all relays.
- Both loaded at startup from JSON files specified in `.env`.

### DDNS

Built-in dynamic DNS updater for home connections with changing IPs. Runs as a background goroutine alongside WoT refresh and cloud backups. Detects public IP via external services (ipify, icanhazip, ifconfig.me), caches last known IP, and only calls the provider API when the IP changes. Configured via `DDNS_*` env vars in `.env`. Providers: Cloudflare (API token + zone/record IDs), DuckDNS (token), No-IP (username/password). Set `DDNS_PROVIDER="none"` to disable.

## Configuration

All config via `.env` file (loaded by godotenv). See `.env.example` for full reference. Key relay list files:
- `relays_import.json` — Seed relays for fetching notes
- `relays_blastr.json` — Target relays for publishing outbox events
- `whitelisted_npubs.json` / `blacklisted_npubs.json` — Access control

## Utilities

- **republish_npubs.sh** — Fetches profiles (kind 0), notes (kind 1), contact lists (kind 3), and relay lists (kind 10002) for whitelisted npubs from source relays and republishes them to all blastr relays. Uses `nak` CLI (`~/.local/bin/nak`). Run to improve discoverability of whitelisted npubs across the relay network.

## Fork Management

This is a fork of `bitvora/haven`. Upstream is tracked as the `upstream` remote:
```bash
git fetch upstream
git merge upstream/master
```
Upstream CI workflow files are excluded (GitHub token lacks `workflow` scope).

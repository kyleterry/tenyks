# tenyks

An IRC bot written in Go. Tenyks relays IRC messages to external service
clients over a bidirectional gRPC stream secured with mutual TLS (mTLS).
Services authenticate using client certificates that encode the channel paths
they are permitted to access.

## Architecture

```
IRC ──► tenyks (gRPC server) ──► service clients (gRPC stream, mTLS)
                ▲                        │
                └────────────────────────┘
```

- **tenyks** — connects to IRC, fans messages out to every registered service
  client, and routes replies back to the appropriate channel.
- **service clients** — long-lived processes that connect to tenyks, receive
  matched messages, and optionally send replies.
- **tenyksctl** — administration CLI for issuing client certificates.

## Getting started

### Prerequisites

```bash
nix develop   # enter the dev shell (Go, protoc, air, etc.)
```

### Run

```bash
go run ./cmd/tenyks
```

### Build

```bash
go build ./cmd/tenyks ./cmd/tenyksctl
```

### Test

```bash
go test ./...
```

### Live reload

```bash
air
```

## Service client certificates

Tenyks requires every service client to present a valid mTLS certificate signed
by the same CA as the server. Certificates embed a custom X.509 extension that
encodes the destination paths the client is allowed to access.

Use `tenyksctl generate-client-certificate` to issue certificates.

### Basic usage (files written to disk)

```bash
tenyksctl generate-client-certificate \
  -ca-cert  ca.crt \
  -ca-key   ca.key \
  -name     weather-service \
  -paths    "libera/#weather,libera/#general" \
  -days     365
```

Writes `weather-service.crt` and `weather-service.key` to the current directory.

### Encrypted bundle for safe delivery

When issuing a certificate for someone else, use `-bundle` to encrypt the
certificate, private key, and CA cert into a single age-encrypted archive. The
private key never needs to travel in plaintext.

**Step 1 — recipient generates an age keypair (one time):**

```bash
age-keygen -o key.txt
# Public key: age1abc123...
```

The recipient shares only the public key (`age1abc123...`) with you.

**Step 2 — issue and encrypt the certificate:**

```bash
tenyksctl generate-client-certificate \
  -ca-cert        ca.crt \
  -ca-key         ca.key \
  -name           weather-service \
  -paths          "libera/#weather" \
  -bundle \
  -age-public-key age1abc123...
```

Writes `weather-service.tar.gz.age`. Send this file to the recipient over any
channel — it is safe to share publicly.

**Step 3 — recipient decrypts:**

```bash
age -d -i key.txt weather-service.tar.gz.age | tar xz
```

Produces three files:

| File | Description |
|---|---|
| `weather-service.crt` | Client certificate |
| `weather-service.key` | Private key (mode 0600) |
| `ca.crt` | CA certificate for verifying the server |

### All flags

| Flag | Default | Description |
|---|---|---|
| `-ca-cert` | — | Path to CA certificate (required) |
| `-ca-key` | — | Path to CA private key (required) |
| `-name` | — | Service name; used as the certificate CN (required) |
| `-paths` | (all) | Comma-separated allowed destination paths |
| `-days` | 365 | Certificate validity period in days |
| `-bundle` | false | Produce an age-encrypted bundle |
| `-age-public-key` | — | Recipient age public key (required with `-bundle`) |
| `-out` | `<name>.tar.gz.age` | Bundle output path (with `-bundle`) |
| `-out-cert` | `<name>.crt` | Certificate output path (without `-bundle`) |
| `-out-key` | `<name>.key` | Private key output path (without `-bundle`) |

### Path matching

Paths encode which IRC server+channel combinations a service may receive
messages from. Matching rules:

| Path value | Matches |
|---|---|
| `libera/#general` | Exactly that channel on that server |
| `libera` | All channels on the libera server |
| _(empty)_ | All paths on all servers |

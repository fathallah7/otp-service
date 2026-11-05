# OTP Service

**A one-time password (OTP) API with asynchronous delivery, built with Go, Gin, Redis, and Asynq.**

[![Go](https://img.shields.io/badge/Go-Backend-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![Gin](https://img.shields.io/badge/Gin-HTTP%20API-008ECF)](https://gin-gonic.com/)
[![Redis](https://img.shields.io/badge/Redis-Storage%20%26%20Queue-DC382D?logo=redis&logoColor=white)](https://redis.io/)
![Status](https://img.shields.io/badge/status-in%20development-F0AD4E)

OTP Service lets clients request and verify six-digit codes through **email** or **WhatsApp**. The HTTP API validates recipients, enforces request limits, and queues delivery jobs. A separate worker generates each code, stores its HMAC-SHA256 digest in Redis, and sends it through the selected provider. Verification atomically consumes a valid code so it cannot be reused.

## Features

| Capability | Current behavior |
| --- | --- |
| Recipient validation | Email addresses and international phone numbers |
| Normalization | Lowercase email domains and E.164 phone numbers |
| OTP generation | Six digits, including possible leading zeros, using `crypto/rand` |
| OTP protection | HMAC-SHA256 digests and recipient-derived Redis keys |
| Expiration | Five minutes from storage by the worker |
| Background delivery | Redis-backed Asynq queue with a separate worker process |
| Worker configuration | Five concurrent tasks; up to three retries per failed task |
| Email delivery | SMTP with mandatory TLS and PLAIN authentication |
| WhatsApp delivery | Meta Cloud API integration with a development test payload |
| Verification | Atomic digest comparison and single-use consumption via Redis Lua |
| Verification attempts | Code invalidated after five incorrect, well-formed guesses |
| Request limits | Per recipient: 60-second cooldown and five requests per one-hour window |
| Tests | Recipient validation unit tests |

## Architecture

```text
Client -> Gin API -> Validate recipient -> Check request limits -> Asynq queue (Redis)
                                                                      |
                                                               OTP worker
                                                                      |
                                                         Generate six-digit code
                                                                      |
                                                     Store HMAC digest (5-minute TTL)
                                                                      |
                                                         Email / WhatsApp provider

Client -> Gin API -> Validate recipient and code -> Redis Lua: verify and consume
```

`POST /api/v1/otp/request` returns after the job is enqueued. The API does not wait for the delivery provider. The API and worker must use the same Redis instance and `OTP_HMAC_KEY`.

The queue is named `otp` and handles tasks of type `otp:deliver`. Each task contains an internal request ID and the normalized recipient; it contains no OTP. Codes are generated during worker execution. Request IDs are logged by the worker but are not returned by the API.

## Project structure

```text
otp-service/
├── cmd/
│   ├── api/main.go          # HTTP API startup and dependency wiring
│   └── worker/main.go       # Delivery worker startup
├── internal/
│   ├── config/              # Environment configuration
│   ├── delivery/            # Dispatcher, SMTP sender, WhatsApp sender
│   ├── otp/                 # DTOs, handlers, service, validation, HMAC, Redis scripts
│   ├── queue/               # Task payload, producer, handler, worker configuration
│   ├── server/              # Gin routes and centralized error responses
│   └── storage/             # Redis connection initialization
├── .env.example
├── go.mod
└── go.sum
```

Key dependencies are [Gin](https://github.com/gin-gonic/gin), [go-redis](https://github.com/redis/go-redis), [Asynq](https://github.com/hibiken/asynq), [phonenumbers](https://github.com/nyaruka/phonenumbers), [go-mail](https://github.com/wneessen/go-mail), and [godotenv](https://github.com/joho/godotenv). WhatsApp requests use Go's `net/http`.

## Getting started

### Prerequisites

- Go **1.27.1 or later**, as declared in `go.mod`.
- A running Redis instance reachable by both the API and worker.
- SMTP credentials for email delivery, or WhatsApp Cloud API credentials and a template matching the development payload for WhatsApp testing.
- OpenSSL if using the secret-generation command below.

### 1. Install dependencies and check Redis

Run commands from the repository root:

```bash
go mod download
redis-cli -h localhost -p 6379 ping
# PONG
```

Use the host and port of your Redis instance if it is not local. The current Redis clients configure an address only; Redis authentication, TLS, and database selection are not exposed through environment variables.

### 2. Configure the environment

For a new local setup, copy the template:

```bash
cp .env.example .env
```

Both processes load `.env` from their working directory. Existing environment variables take precedence. `.env` is ignored by Git.

| Variable | Default | Purpose |
| --- | --- | --- |
| `APP_ENV` | `development` | `production` enables Gin release mode |
| `PORT` | `8080` | HTTP API port |
| `REDIS_ADDR` | `localhost:6379` | Redis address for OTP records, limits, and queued jobs |
| `OTP_HMAC_KEY` | None | Required by both processes: exactly 32 bytes encoded as 64 hexadecimal characters |
| `SMTP_HOST` | None | SMTP server hostname |
| `SMTP_PORT` | None | SMTP server port |
| `SMTP_USERNAME` | None | SMTP authentication username |
| `SMTP_PASSWORD` | None | SMTP authentication password |
| `SMTP_FROM` | None | Sender email address |
| `WHATSAPP_API_VERSION` | None | Graph API version used in the request URL |
| `WHATSAPP_PHONE_NUMBER_ID` | None | Sender phone-number ID |
| `WHATSAPP_ACCESS_TOKEN` | None | Bearer token for WhatsApp requests |
| `WHATSAPP_TEMPLATE_NAME` | None | Development template name |
| `WHATSAPP_TEMPLATE_LANGUAGE` | None | Template language code, such as `en_US` |

`REDIS_ADDR` is not currently listed in `.env.example`; add it to `.env` if you need a different address. Configure the provider variables for the channel you plan to use.

Generate the HMAC key and copy the output into `OTP_HMAC_KEY`:

```bash
openssl rand -hex 32
```

Use the same key in both processes and keep it stable across restarts. Changing it makes existing OTP records inaccessible and changes the keys used for request limits. Keep credentials and the HMAC key out of version control.

The WhatsApp sender currently builds a test template body with three text parameters: `Test User`, the generated OTP, and a hard-coded date. The configured template must match that payload. A production authentication-template payload is not implemented.

### 3. Start the API and worker

In one terminal:

```bash
go run ./cmd/api
```

In another terminal, from the same repository root:

```bash
go run ./cmd/worker
```

The API listens on `http://localhost:8080` by default. Both processes check Redis connectivity and validate the HMAC key at startup. Without a running worker, the API can still accept requests, but delivery jobs remain queued.

### 4. Check the API

```bash
curl -i http://localhost:8080/health
```

Expected response: `200 OK`.

```json
{"status":"ok"}
```

This endpoint reports that the HTTP server is responding. It does not check current Redis connectivity, worker availability, or delivery providers.

## API reference

The routes currently require no API authentication. Request and verification bodies use JSON.

### `POST /api/v1/otp/request`

| Field | Type | Requirement |
| --- | --- | --- |
| `recipient` | String | Required; at most 254 characters; an email address or valid international phone number beginning with `+` |

The recipient determines the channel. Surrounding whitespace is trimmed; email domains are lowercased while the local part keeps its case. Phone numbers are normalized to E.164.

**Email:**

```bash
curl -i -X POST http://localhost:8080/api/v1/otp/request \
  -H 'Content-Type: application/json' \
  -d '{"recipient":"user@example.com"}'
```

**WhatsApp:** replace the placeholder with your international test number.

```bash
curl -i -X POST http://localhost:8080/api/v1/otp/request \
  -H 'Content-Type: application/json' \
  -d '{"recipient":"+YOUR_E164_TEST_NUMBER"}'
```

**Success — `202 Accepted`:**

```json
{"message":"OTP request accepted"}
```

This confirms that the job was queued. It does not confirm that a code has been generated or delivered. Provider failures are handled by the worker and cannot change this response. No delivery-status endpoint is available.

Request limits are shared through Redis and apply per normalized recipient/channel:

- A **60-second cooldown** follows each request that passes the limit check.
- At most **five requests** can pass the check within a **fixed one-hour window**, starting with the first counted request.
- The cooldown check runs first. Rejected requests do not extend either window.
- Limits are reserved before enqueueing; an enqueue failure currently still consumes the cooldown and a request allowance.

**Cooldown — `429 Too Many Requests`:**

```json
{
  "error": {
    "code": "resend_cooldown",
    "message": "Please wait before requesting another OTP"
  }
}
```

### `POST /api/v1/otp/verify`

| Field | Type | Requirement |
| --- | --- | --- |
| `recipient` | String | Required; same validation and normalization as the request endpoint |
| `code` | String | Required; exactly six ASCII digits; preserve leading zeros |

Use the code received from the delivery provider:

```bash
curl -i -X POST http://localhost:8080/api/v1/otp/verify \
  -H 'Content-Type: application/json' \
  -d '{"recipient":"user@example.com","code":"012345"}'
```

**Success — `200 OK`:**

```json
{"message":"OTP verified successfully"}
```

Successful verification atomically deletes the digest and failed-attempt counter. A second verification with the same code fails. Five incorrect, well-formed guesses invalidate the stored code; malformed codes return `400` without incrementing the counter. Storing a new code resets the failed-attempt counter.

**Incorrect, expired, consumed, missing, or attempt-limited code — `422 Unprocessable Entity`:**

```json
{"error":"invalid or expired OTP"}
```

This response also applies while delivery is queued if no valid code is already stored. Its `error` field is a string; centralized errors below use an object.

### Error responses

Centralized errors have this shape:

```json
{
  "error": {
    "code": "invalid_recipient",
    "message": "Invalid email or phone number"
  }
}
```

| Status | Code | Meaning |
| --- | --- | --- |
| `400` | `invalid_request` | Invalid JSON, missing/invalid fields, or a malformed verification code |
| `400` | `invalid_recipient` | Recipient is not a valid email address or international phone number |
| `429` | `resend_cooldown` | Request made during the recipient's 60-second cooldown |
| `429` | `request_rate_limit` | Recipient has reached the five-request hourly limit |
| `500` | `internal_error` | Unexpected failure, including Redis or enqueue errors |

The hourly-limit message is `too many OTP requests, please try again later`. Rate-limit responses currently omit a `Retry-After` header. The centralized handler also defines `502 delivery_failed`, but provider delivery runs in the worker, so provider failures are not returned by the current request or verification routes.

## Storage and delivery behavior

- **OTP records:** Redis stores HMAC-SHA256 digests under `otp:request:<digest>`, with a five-minute TTL starting when the worker stores the code, before calling the provider. The key includes the channel and normalized recipient in its HMAC input.
- **Related records:** `:attempts`, `:cooldown`, and `:hourly` suffixes track failed guesses and request limits. Redis Lua scripts make saving/resetting attempts, verifying/consuming, and checking request limits atomic within each operation.
- **Queue contents:** Job payloads contain the normalized recipient in plaintext and an internal request ID. The OTP itself is not stored in the queue. HMAC-derived OTP keys do not conceal recipients in queued payloads.
- **Retries:** Failed jobs are configured for up to three retries. Each execution generates a new code and overwrites that recipient's previous digest, restarting its TTL and attempt counter. Malformed JSON payloads and missing required task fields are skipped without retrying.

### Current limitations

- Delivery failures leave the stored digest until it expires or is overwritten; there is no cleanup tied to failed delivery.
- Request IDs do not enforce idempotency or ordering. Overlapping jobs and retries can replace a code that was already delivered, and there is no guarantee of exactly-once delivery.
- Rate-limit reservation and job enqueueing are separate operations, with no rollback if enqueueing fails.
- API authentication and per-IP/global abuse controls are not implemented.
- Worker logs report provider-call success or task failure; delivery receipts and a client-facing job-status endpoint are not implemented.
- WhatsApp uses a development test payload. Production authentication-template support remains to be implemented.
- Automated coverage currently consists of recipient validation tests; queue, handler, Redis-script, and delivery integration tests remain to be added.

## Development commands

```bash
# Format source code
go fmt ./...

# Compile the API, worker, and internal packages
go build ./...

# Run the existing unit tests
go test ./...
```

Inspect OTP-related records in a local development Redis instance:

```bash
redis-cli --scan --pattern 'otp:request:*'
redis-cli TTL 'otp:request:<digest>'
```

Replace `<digest>` with a key suffix returned by the scan. The scan also includes request-limit and attempt-counter keys.


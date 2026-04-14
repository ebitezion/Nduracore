# Alchemy Webhook Integration

## Endpoint
`POST /v1/wallets/:id/deposits/webhook`

Example:
`POST /v1/wallets/8c1f2d5d-8b6a-4e3f-ae91-7f5b47f3a301/deposits/webhook`

## Required Headers
- `X-Tenant-ID`: tenant context (must match wallet tenant).
- `X-Alchemy-Signature`: HMAC-SHA256 signature of raw request body.
- `Content-Type: application/json`

## Signature Verification
If `ALCHEMY_WEBHOOK_SIGNING_SECRET` is configured, requests are verified as:
1. Compute HMAC SHA256 over raw request body using `ALCHEMY_WEBHOOK_SIGNING_SECRET`.
2. Compare hex digest with `X-Alchemy-Signature` header.
3. Header may be plain hex or prefixed with `sha256=`.

## Supported Payload Shapes

### 1) Simple payload
```json
{
  "tx_hash": "0xabc123...",
  "amount_minor": 250000,
  "confirmations": 12,
  "status": "confirmed"
}
```

### 2) Alchemy activity payload
```json
{
  "event": {
    "activity": [
      {
        "hash": "0xabc123...",
        "value": "0.25",
        "numConfirmations": 12,
        "status": "confirmed",
        "rawContract": {
          "value": ""
        }
      },
      {
        "hash": "0xdef456...",
        "rawContract": {
          "value": "0x3e8"
        },
        "confirmations": 3
      }
    ]
  }
}
```

Notes:
- `rawContract.value` hex is prioritized for `amount_minor` parsing.
- If `rawContract.value` is empty, `value` is interpreted and converted.
- If `status` is absent, status defaults to:
  - `confirmed` when confirmations > 0
  - `pending` when confirmations == 0

## Success Response
`202 Accepted`

```json
{
  "deposits": [
    {
      "tx_hash": "0xabc123...",
      "amount_minor": 250000,
      "confirmations": 12,
      "status": "confirmed"
    }
  ]
}
```

## Error Responses
- `400 Bad Request`
  - Missing `X-Tenant-ID`
  - Invalid payload
  - No deposit activity in payload
- `401 Unauthorized`
  - Invalid webhook signature
- `404 Not Found`
  - Unknown wallet ID for provided tenant

## Environment Variables
- `ALCHEMY_WEBHOOK_SIGNING_SECRET`
- `ALCHEMY_WEBHOOK_SIGNING_SECRET_FILE`
- `ALCHEMY_WEBHOOK_SIGNING_SECRET_REF`

## Example Signature Generation
```bash
body='{"tx_hash":"0xabc","amount_minor":250000,"confirmations":12,"status":"confirmed"}'
secret='replace-with-webhook-secret'
signature=$(printf '%s' "$body" | openssl dgst -sha256 -hmac "$secret" -binary | xxd -p -c 256)

curl -X POST "http://localhost:4000/v1/wallets/<wallet-id>/deposits/webhook" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: tenant-1" \
  -H "X-Alchemy-Signature: $signature" \
  -d "$body"
```

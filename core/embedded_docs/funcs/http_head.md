# http_head

Sends an HTTP HEAD to `url` and returns the response as a map. The server returns headers without a body; the response map's `body` is omitted.

```rad
http_head(url: str, *, body: any?, json: any?, headers: map?, insecure: bool = false) -> { "success": bool, "status_code"?: int, "headers": map, "body"?: any, "error"?: str, "duration_seconds": float }
```

```rad
r = http_head("https://api.example.com/resource")
if r.success:
    print(r.body)
```

## Notes

**Response map keys:**

- `success: bool` - whether the request succeeded.
- `duration_seconds: float` - total request time.
- `status_code?: int` - present when a response was received.
- `headers: map` - response headers.
- `body?: any` - response body, JSON-decoded when possible.
- `error?: str` - error message when `success` is false.

**Body vs JSON:** `body` is sent as-is; `json` is JSON-serialised and sets `Content-Type: application/json` when no headers are supplied. The two are mutually exclusive.

**Insecure:** pass `insecure=true` to skip TLS certificate verification.

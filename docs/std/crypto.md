# std.crypto

`std.crypto` provides hashing, HMAC, JWT helpers, and password hashing APIs.

## Modules

- `std.crypto`
- `std.crypto.jwt`
- `std.crypto.password`

## std.crypto API

| Function | Parameters | Returns | Notes |
| --- | --- | --- | --- |
| `sha256` | `data | bytes` | `bytes` | SHA-256 digest |
| `sha512` | `data | bytes` | `bytes` | SHA-512 digest |
| `hmacSha256` | `key | bytes`, `data | bytes` | `bytes` | HMAC-SHA256 signature |
| `verifyHmacSha256` | `key | bytes`, `data | bytes`, `signature | bytes` | `bool` | constant-time compare |
| `base64UrlEncode` | `data | bytes` | `string` | no padding |
| `base64UrlDecode` | `text | string` | `bytes` | URL-safe Base64 decode |

## std.crypto.jwt API

### Sign

| Function | Parameters | Returns |
| --- | --- | --- |
| `signHS256` | `payload | dict<any>`, `secret | bytes` | `string` |
| `signRS256` | `payload | dict<any>`, `privateKeyPem | bytes` | `string` |
| `signES256` | `payload | dict<any>`, `privateKeyPem | bytes` | `string` |

`sign*` sets `alg` and `typ` in JWT header automatically.

### Verify

| Function | Parameters | Returns |
| --- | --- | --- |
| `verifyHS256` | `token | string`, `secret | bytes` | `dict<any>` |
| `verifyHS256At` | `token | string`, `secret | bytes`, `nowUnixSeconds | int` | `dict<any>` |
| `verifyRS256` | `token | string`, `publicKeyPem | bytes` | `dict<any>` |
| `verifyRS256At` | `token | string`, `publicKeyPem | bytes`, `nowUnixSeconds | int` | `dict<any>` |
| `verifyES256` | `token | string`, `publicKeyPem | bytes` | `dict<any>` |
| `verifyES256At` | `token | string`, `publicKeyPem | bytes`, `nowUnixSeconds | int` | `dict<any>` |

Verify returns status object:

```avenir
{
    valid: bool,
    header?: dict<any>,
    payload?: dict<any>,
    reason?: string,
    error?: string
}
```

### Claim checks

When `exp`, `nbf`, `iat` are present and numeric:

- `exp`: token must not be expired
- `nbf`: token must already be active
- `iat`: token issue time cannot be too far in the future

## std.crypto.password API

| Function | Parameters | Returns |
| --- | --- | --- |
| `hash` | `password | string` | `string` |
| `verify` | `password | string`, `encoded | string` | `bool` |

Current implementation uses bcrypt in runtime builtins.

## Security notes

- Prefer sufficiently long random secrets for HS256.
- Keep private keys private and use trusted PEM key material only.
- Always check `status.valid` before trusting `payload`.
- Do not log raw passwords or JWT secrets.

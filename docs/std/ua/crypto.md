# std.crypto — Криптографія

Модуль `std.crypto` надає хешування, HMAC, JWT-хелпери та хешування паролів.

## Імпорт

```avenir
import std.crypto;
import std.crypto.jwt;
import std.crypto.password;
```

## std.crypto — Хешування та HMAC

| Функція | Опис |
|---------|------|
| `sha256(data \| string) \| string` | SHA-256 хеш (hex) |
| `sha512(data \| string) \| string` | SHA-512 хеш (hex) |
| `hmacSha256(data \| string, key \| string) \| string` | HMAC-SHA256 (hex) |
| `verifyHmacSha256(data \| string, key \| string, sig \| string) \| bool` | Перевірка HMAC-SHA256 |
| `base64UrlEncode(data \| bytes) \| string` | Base64URL кодування |
| `base64UrlDecode(data \| string) \| bytes` | Base64URL декодування |
| `randomBytes(n \| int) \| bytes` | Криптографічно безпечні випадкові байти (1-1024) |

## std.crypto.jwt — JSON Web Tokens

| Функція | Опис |
|---------|------|
| `sign(claims \| dict<any>, secret \| string) \| string` | Підписати JWT |
| `verify(token \| string, secret \| string) \| dict<any>` | Перевірити та декодувати JWT |

### Перевірки claims

| Функція | Опис |
|---------|------|
| `isExpired(claims \| dict<any>) \| bool` | Перевірити закінчення терміну |
| `getSubject(claims \| dict<any>) \| string` | Отримати subject |
| `getIssuer(claims \| dict<any>) \| string` | Отримати issuer |

### Приклад JWT

```avenir
import std.crypto.jwt;

fun main() | void {
    var claims | dict<any> = {
        sub: "user123",
        iss: "myapp",
        exp: 1700000000
    };

    var token | string = std.crypto.jwt.sign(claims, "my-secret");
    print(token);

    var decoded | dict<any> = std.crypto.jwt.verify(token, "my-secret");
    print(decoded.get("sub"));
}
```

## std.crypto.password — Хешування паролів

| Функція | Опис |
|---------|------|
| `hash(password \| string) \| string` | Хешування пароля (bcrypt) |
| `verify(password \| string, hash \| string) \| bool` | Перевірка пароля |

### Приклад паролів

```avenir
import std.crypto.password;

fun main() | void {
    var hashed | string = std.crypto.password.hash("mypassword");
    print(hashed);

    var valid | bool = std.crypto.password.verify("mypassword", hashed);
    print(valid);
}
```

## Примітки безпеки

- JWT використовує HMAC-SHA256 для підпису
- Хешування паролів використовує bcrypt з безпечними параметрами за замовчуванням
- `randomBytes` використовує `crypto/rand` (CSPRNG)
- HMAC-перевірка використовує порівняння з постійним часом

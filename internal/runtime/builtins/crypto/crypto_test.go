package crypto_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/x509"
	"encoding/pem"
	"testing"
	"time"

	"avenir/internal/runtime"
	"avenir/internal/runtime/builtins"
	"avenir/internal/value"
)

func callBuiltin(t *testing.T, env *runtime.Env, name string, args ...value.Value) (value.Value, error) {
	t.Helper()
	b := builtins.LookupByName(name)
	if b == nil {
		t.Fatalf("builtin %q not found", name)
	}
	argsIface := make([]interface{}, len(args))
	for i, arg := range args {
		argsIface[i] = arg
	}
	res, err := b.Call(env, argsIface)
	if err != nil {
		return value.Value{}, err
	}
	val, ok := res.(value.Value)
	if !ok {
		t.Fatalf("builtin %q returned non-value %T", name, res)
	}
	return val, nil
}

func TestSHA256AndSHA512(t *testing.T) {
	env := runtime.DefaultEnv()
	data := []byte("abc")

	s256, err := callBuiltin(t, env, "__builtin_crypto_sha256", value.Bytes(data))
	if err != nil {
		t.Fatalf("sha256 error: %v", err)
	}
	if s256.Kind != value.KindBytes {
		t.Fatalf("sha256 expected bytes, got %v", s256.Kind)
	}
	want256 := sha256.Sum256(data)
	if string(s256.Bytes) != string(want256[:]) {
		t.Fatalf("sha256 mismatch")
	}

	s512, err := callBuiltin(t, env, "__builtin_crypto_sha512", value.Bytes(data))
	if err != nil {
		t.Fatalf("sha512 error: %v", err)
	}
	if s512.Kind != value.KindBytes {
		t.Fatalf("sha512 expected bytes, got %v", s512.Kind)
	}
	want512 := sha512.Sum512(data)
	if string(s512.Bytes) != string(want512[:]) {
		t.Fatalf("sha512 mismatch")
	}
}

func TestHMACSHA256(t *testing.T) {
	env := runtime.DefaultEnv()
	key := value.Bytes([]byte("secret"))
	data := value.Bytes([]byte("payload"))

	sig, err := callBuiltin(t, env, "__builtin_crypto_hmac_sha256", key, data)
	if err != nil {
		t.Fatalf("hmac error: %v", err)
	}
	if sig.Kind != value.KindBytes {
		t.Fatalf("expected bytes signature")
	}

	ok, err := callBuiltin(t, env, "__builtin_crypto_hmac_sha256_verify", key, data, sig)
	if err != nil {
		t.Fatalf("hmac verify error: %v", err)
	}
	if ok.Kind != value.KindBool || !ok.Bool {
		t.Fatalf("expected valid signature")
	}

	bad, err := callBuiltin(t, env, "__builtin_crypto_hmac_sha256_verify", key, data, value.Bytes([]byte("wrong")))
	if err != nil {
		t.Fatalf("hmac verify bad error: %v", err)
	}
	if bad.Kind != value.KindBool || bad.Bool {
		t.Fatalf("expected invalid signature")
	}
}

func TestBase64URLRoundTrip(t *testing.T) {
	env := runtime.DefaultEnv()
	data := value.Bytes([]byte{1, 2, 3, 4, 5, 255})
	enc, err := callBuiltin(t, env, "__builtin_crypto_base64url_encode", data)
	if err != nil {
		t.Fatalf("encode error: %v", err)
	}
	if enc.Kind != value.KindString {
		t.Fatalf("expected string")
	}
	dec, err := callBuiltin(t, env, "__builtin_crypto_base64url_decode", enc)
	if err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if dec.Kind != value.KindBytes || string(dec.Bytes) != string(data.Bytes) {
		t.Fatalf("roundtrip mismatch")
	}
}

func TestJWTHS256SignVerify(t *testing.T) {
	env := runtime.DefaultEnv()
	now := time.Now().Unix()
	payload := value.Dict(map[string]value.Value{
		"sub": value.Str("user-1"),
		"exp": value.Int(now + 60),
	})
	token, err := callBuiltin(t, env, "__builtin_crypto_jwt_sign_hs256", value.Dict(map[string]value.Value{}), payload, value.Bytes([]byte("top-secret")))
	if err != nil {
		t.Fatalf("jwt sign hs256 error: %v", err)
	}
	status, err := callBuiltin(t, env, "__builtin_crypto_jwt_verify_hs256", token, value.Bytes([]byte("top-secret")), value.Int(now))
	if err != nil {
		t.Fatalf("jwt verify hs256 error: %v", err)
	}
	if status.Kind != value.KindDict || !status.Dict["valid"].Bool {
		t.Fatalf("expected valid JWT status")
	}
}

func TestJWTHS256Expired(t *testing.T) {
	env := runtime.DefaultEnv()
	now := time.Now().Unix()
	payload := value.Dict(map[string]value.Value{
		"sub": value.Str("user-1"),
		"exp": value.Int(now - 1),
	})
	token, err := callBuiltin(t, env, "__builtin_crypto_jwt_sign_hs256", value.Dict(map[string]value.Value{}), payload, value.Bytes([]byte("top-secret")))
	if err != nil {
		t.Fatalf("jwt sign hs256 error: %v", err)
	}
	status, err := callBuiltin(t, env, "__builtin_crypto_jwt_verify_hs256", token, value.Bytes([]byte("top-secret")), value.Int(now))
	if err != nil {
		t.Fatalf("jwt verify hs256 error: %v", err)
	}
	if status.Kind != value.KindDict || status.Dict["valid"].Bool {
		t.Fatalf("expected invalid JWT status")
	}
	if status.Dict["reason"].Kind != value.KindString || status.Dict["reason"].Str != "expired" {
		t.Fatalf("expected expired reason")
	}
}

func TestJWTRS256AndES256(t *testing.T) {
	env := runtime.DefaultEnv()
	now := time.Now().Unix()
	payload := value.Dict(map[string]value.Value{
		"sub": value.Str("user-2"),
		"exp": value.Int(now + 60),
	})

	rsaPriv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("rsa key gen: %v", err)
	}
	rsaPrivPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(rsaPriv)})
	rsaPubDER, err := x509.MarshalPKIXPublicKey(&rsaPriv.PublicKey)
	if err != nil {
		t.Fatalf("rsa pub marshal: %v", err)
	}
	rsaPubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: rsaPubDER})

	rsaToken, err := callBuiltin(t, env, "__builtin_crypto_jwt_sign_rs256", value.Dict(map[string]value.Value{}), payload, value.Bytes(rsaPrivPEM))
	if err != nil {
		t.Fatalf("jwt sign rs256: %v", err)
	}
	rsaStatus, err := callBuiltin(t, env, "__builtin_crypto_jwt_verify_rs256", rsaToken, value.Bytes(rsaPubPEM), value.Int(now))
	if err != nil {
		t.Fatalf("jwt verify rs256: %v", err)
	}
	if rsaStatus.Kind != value.KindDict || !rsaStatus.Dict["valid"].Bool {
		t.Fatalf("expected valid RS256 status")
	}

	esPriv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("ecdsa key gen: %v", err)
	}
	esPrivDER, err := x509.MarshalECPrivateKey(esPriv)
	if err != nil {
		t.Fatalf("ecdsa priv marshal: %v", err)
	}
	esPrivPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: esPrivDER})
	esPubDER, err := x509.MarshalPKIXPublicKey(&esPriv.PublicKey)
	if err != nil {
		t.Fatalf("ecdsa pub marshal: %v", err)
	}
	esPubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: esPubDER})

	esToken, err := callBuiltin(t, env, "__builtin_crypto_jwt_sign_es256", value.Dict(map[string]value.Value{}), payload, value.Bytes(esPrivPEM))
	if err != nil {
		t.Fatalf("jwt sign es256: %v", err)
	}
	esStatus, err := callBuiltin(t, env, "__builtin_crypto_jwt_verify_es256", esToken, value.Bytes(esPubPEM), value.Int(now))
	if err != nil {
		t.Fatalf("jwt verify es256: %v", err)
	}
	if esStatus.Kind != value.KindDict || !esStatus.Dict["valid"].Bool {
		t.Fatalf("expected valid ES256 status")
	}
}

func TestPasswordHashVerify(t *testing.T) {
	env := runtime.DefaultEnv()
	hash, err := callBuiltin(t, env, "__builtin_crypto_password_hash", value.Str("pa$$w0rd"))
	if err != nil {
		t.Fatalf("password hash error: %v", err)
	}
	if hash.Kind != value.KindString || hash.Str == "" {
		t.Fatalf("expected non-empty hash")
	}
	ok, err := callBuiltin(t, env, "__builtin_crypto_password_verify", value.Str("pa$$w0rd"), hash)
	if err != nil {
		t.Fatalf("password verify error: %v", err)
	}
	if ok.Kind != value.KindBool || !ok.Bool {
		t.Fatalf("expected password to verify")
	}
	bad, err := callBuiltin(t, env, "__builtin_crypto_password_verify", value.Str("wrong"), hash)
	if err != nil {
		t.Fatalf("password verify wrong error: %v", err)
	}
	if bad.Kind != value.KindBool || bad.Bool {
		t.Fatalf("expected wrong password to fail")
	}
}

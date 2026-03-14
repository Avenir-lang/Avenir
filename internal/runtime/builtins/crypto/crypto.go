package crypto

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/subtle"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"math"
	"math/big"
	"strings"

	"avenir/internal/runtime/builtins"
	"avenir/internal/value"

	"golang.org/x/crypto/bcrypt"
)

var jwtBase64 = base64.RawURLEncoding

func init() {
	registerSHA256()
	registerSHA512()
	registerHMACSHA256()
	registerHMACSHA256Verify()
	registerBase64URLEncode()
	registerBase64URLDecode()
	registerJWTSignHS256()
	registerJWTVerifyHS256()
	registerJWTSignRS256()
	registerJWTVerifyRS256()
	registerJWTSignES256()
	registerJWTVerifyES256()
	registerPasswordHash()
	registerPasswordVerify()
	registerRandomBytes()
}

func registerSHA256() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:           builtins.CryptoSHA256,
			Name:         "__builtin_crypto_sha256",
			Arity:        1,
			ParamNames:   []string{"data"},
			Params:       []builtins.TypeRef{{Kind: builtins.TypeBytes}},
			Result:       builtins.TypeRef{Kind: builtins.TypeBytes},
			ReceiverType: builtins.TypeVoid,
			MethodName:   "",
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			data, err := requireBytesArg(args, 0, "__builtin_crypto_sha256")
			if err != nil {
				return value.Value{}, err
			}
			sum := sha256.Sum256(data)
			return value.Bytes(sum[:]), nil
		},
	})
}

func registerSHA512() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:           builtins.CryptoSHA512,
			Name:         "__builtin_crypto_sha512",
			Arity:        1,
			ParamNames:   []string{"data"},
			Params:       []builtins.TypeRef{{Kind: builtins.TypeBytes}},
			Result:       builtins.TypeRef{Kind: builtins.TypeBytes},
			ReceiverType: builtins.TypeVoid,
			MethodName:   "",
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			data, err := requireBytesArg(args, 0, "__builtin_crypto_sha512")
			if err != nil {
				return value.Value{}, err
			}
			sum := sha512.Sum512(data)
			return value.Bytes(sum[:]), nil
		},
	})
}

func registerHMACSHA256() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:           builtins.CryptoHMACSHA256,
			Name:         "__builtin_crypto_hmac_sha256",
			Arity:        2,
			ParamNames:   []string{"key", "data"},
			Params:       []builtins.TypeRef{{Kind: builtins.TypeBytes}, {Kind: builtins.TypeBytes}},
			Result:       builtins.TypeRef{Kind: builtins.TypeBytes},
			ReceiverType: builtins.TypeVoid,
			MethodName:   "",
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			key, err := requireBytesArg(args, 0, "__builtin_crypto_hmac_sha256")
			if err != nil {
				return value.Value{}, err
			}
			data, err := requireBytesArg(args, 1, "__builtin_crypto_hmac_sha256")
			if err != nil {
				return value.Value{}, err
			}
			mac := hmac.New(sha256.New, key)
			mac.Write(data)
			return value.Bytes(mac.Sum(nil)), nil
		},
	})
}

func registerHMACSHA256Verify() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:           builtins.CryptoHMACSHA256Verify,
			Name:         "__builtin_crypto_hmac_sha256_verify",
			Arity:        3,
			ParamNames:   []string{"key", "data", "signature"},
			Params:       []builtins.TypeRef{{Kind: builtins.TypeBytes}, {Kind: builtins.TypeBytes}, {Kind: builtins.TypeBytes}},
			Result:       builtins.TypeRef{Kind: builtins.TypeBool},
			ReceiverType: builtins.TypeVoid,
			MethodName:   "",
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			key, err := requireBytesArg(args, 0, "__builtin_crypto_hmac_sha256_verify")
			if err != nil {
				return value.Value{}, err
			}
			data, err := requireBytesArg(args, 1, "__builtin_crypto_hmac_sha256_verify")
			if err != nil {
				return value.Value{}, err
			}
			sig, err := requireBytesArg(args, 2, "__builtin_crypto_hmac_sha256_verify")
			if err != nil {
				return value.Value{}, err
			}
			mac := hmac.New(sha256.New, key)
			mac.Write(data)
			expected := mac.Sum(nil)
			ok := subtle.ConstantTimeCompare(expected, sig) == 1
			return value.Bool(ok), nil
		},
	})
}

func registerBase64URLEncode() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:           builtins.CryptoBase64URLEncode,
			Name:         "__builtin_crypto_base64url_encode",
			Arity:        1,
			ParamNames:   []string{"data"},
			Params:       []builtins.TypeRef{{Kind: builtins.TypeBytes}},
			Result:       builtins.TypeRef{Kind: builtins.TypeString},
			ReceiverType: builtins.TypeVoid,
			MethodName:   "",
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			data, err := requireBytesArg(args, 0, "__builtin_crypto_base64url_encode")
			if err != nil {
				return value.Value{}, err
			}
			return value.Str(jwtBase64.EncodeToString(data)), nil
		},
	})
}

func registerBase64URLDecode() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:           builtins.CryptoBase64URLDecode,
			Name:         "__builtin_crypto_base64url_decode",
			Arity:        1,
			ParamNames:   []string{"text"},
			Params:       []builtins.TypeRef{{Kind: builtins.TypeString}},
			Result:       builtins.TypeRef{Kind: builtins.TypeBytes},
			ReceiverType: builtins.TypeVoid,
			MethodName:   "",
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			text, err := requireStringArg(args, 0, "__builtin_crypto_base64url_decode")
			if err != nil {
				return value.Value{}, err
			}
			decoded, err := jwtBase64.DecodeString(text)
			if err != nil {
				return value.Value{}, fmt.Errorf("invalid base64url data: %w", err)
			}
			return value.Bytes(decoded), nil
		},
	})
}

func registerJWTSignHS256() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:           builtins.CryptoJWTSignHS256,
			Name:         "__builtin_crypto_jwt_sign_hs256",
			Arity:        3,
			ParamNames:   []string{"header", "payload", "secret"},
			Params:       []builtins.TypeRef{{Kind: builtins.TypeDict, Elem: []builtins.TypeRef{{Kind: builtins.TypeAny}}}, {Kind: builtins.TypeDict, Elem: []builtins.TypeRef{{Kind: builtins.TypeAny}}}, {Kind: builtins.TypeBytes}},
			Result:       builtins.TypeRef{Kind: builtins.TypeString},
			ReceiverType: builtins.TypeVoid,
			MethodName:   "",
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			header, payload, signingInput, err := jwtPrepareHeaderPayload(args, "HS256", "__builtin_crypto_jwt_sign_hs256")
			if err != nil {
				return value.Value{}, err
			}
			secret, err := requireBytesArg(args, 2, "__builtin_crypto_jwt_sign_hs256")
			if err != nil {
				return value.Value{}, err
			}
			mac := hmac.New(sha256.New, secret)
			mac.Write([]byte(signingInput))
			sig := mac.Sum(nil)
			_ = header
			_ = payload
			return value.Str(signingInput + "." + jwtBase64.EncodeToString(sig)), nil
		},
	})
}

func registerJWTVerifyHS256() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:           builtins.CryptoJWTVerifyHS256,
			Name:         "__builtin_crypto_jwt_verify_hs256",
			Arity:        3,
			ParamNames:   []string{"token", "secret", "now"},
			Params:       []builtins.TypeRef{{Kind: builtins.TypeString}, {Kind: builtins.TypeBytes}, {Kind: builtins.TypeInt}},
			Result:       builtins.TypeRef{Kind: builtins.TypeDict, Elem: []builtins.TypeRef{{Kind: builtins.TypeAny}}},
			ReceiverType: builtins.TypeVoid,
			MethodName:   "",
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			token, err := requireStringArg(args, 0, "__builtin_crypto_jwt_verify_hs256")
			if err != nil {
				return value.Value{}, err
			}
			secret, err := requireBytesArg(args, 1, "__builtin_crypto_jwt_verify_hs256")
			if err != nil {
				return value.Value{}, err
			}
			nowUnix, err := requireIntArg(args, 2, "__builtin_crypto_jwt_verify_hs256")
			if err != nil {
				return value.Value{}, err
			}
			parts, header, payload, signingInput, signature, status, err := jwtParseForVerify(token)
			if err != nil {
				return value.Value{}, err
			}
			if status != nil {
				return value.Dict(status), nil
			}
			_ = parts
			if alg, ok := header["alg"]; !ok || alg.Kind != value.KindString || alg.Str != "HS256" {
				return value.Dict(jwtStatusInvalid("alg_mismatch", "token alg must be HS256")), nil
			}
			mac := hmac.New(sha256.New, secret)
			mac.Write([]byte(signingInput))
			expected := mac.Sum(nil)
			if subtle.ConstantTimeCompare(expected, signature) != 1 {
				return value.Dict(jwtStatusInvalid("bad_signature", "invalid signature")), nil
			}
			if s := validateJWTClaims(payload, nowUnix); s != nil {
				return value.Dict(s), nil
			}
			return value.Dict(jwtStatusValid(header, payload)), nil
		},
	})
}

func registerJWTSignRS256() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:           builtins.CryptoJWTSignRS256,
			Name:         "__builtin_crypto_jwt_sign_rs256",
			Arity:        3,
			ParamNames:   []string{"header", "payload", "privateKeyPem"},
			Params:       []builtins.TypeRef{{Kind: builtins.TypeDict, Elem: []builtins.TypeRef{{Kind: builtins.TypeAny}}}, {Kind: builtins.TypeDict, Elem: []builtins.TypeRef{{Kind: builtins.TypeAny}}}, {Kind: builtins.TypeBytes}},
			Result:       builtins.TypeRef{Kind: builtins.TypeString},
			ReceiverType: builtins.TypeVoid,
			MethodName:   "",
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			_, _, signingInput, err := jwtPrepareHeaderPayload(args, "RS256", "__builtin_crypto_jwt_sign_rs256")
			if err != nil {
				return value.Value{}, err
			}
			pemBytes, err := requireBytesArg(args, 2, "__builtin_crypto_jwt_sign_rs256")
			if err != nil {
				return value.Value{}, err
			}
			priv, err := parseRSAPrivateKeyPEM(pemBytes)
			if err != nil {
				return value.Value{}, err
			}
			h := sha256.Sum256([]byte(signingInput))
			sig, err := rsa.SignPKCS1v15(rand.Reader, priv, crypto.SHA256, h[:])
			if err != nil {
				return value.Value{}, err
			}
			return value.Str(signingInput + "." + jwtBase64.EncodeToString(sig)), nil
		},
	})
}

func registerJWTVerifyRS256() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:           builtins.CryptoJWTVerifyRS256,
			Name:         "__builtin_crypto_jwt_verify_rs256",
			Arity:        3,
			ParamNames:   []string{"token", "publicKeyPem", "now"},
			Params:       []builtins.TypeRef{{Kind: builtins.TypeString}, {Kind: builtins.TypeBytes}, {Kind: builtins.TypeInt}},
			Result:       builtins.TypeRef{Kind: builtins.TypeDict, Elem: []builtins.TypeRef{{Kind: builtins.TypeAny}}},
			ReceiverType: builtins.TypeVoid,
			MethodName:   "",
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			token, err := requireStringArg(args, 0, "__builtin_crypto_jwt_verify_rs256")
			if err != nil {
				return value.Value{}, err
			}
			pemBytes, err := requireBytesArg(args, 1, "__builtin_crypto_jwt_verify_rs256")
			if err != nil {
				return value.Value{}, err
			}
			nowUnix, err := requireIntArg(args, 2, "__builtin_crypto_jwt_verify_rs256")
			if err != nil {
				return value.Value{}, err
			}
			pub, err := parseRSAPublicKeyPEM(pemBytes)
			if err != nil {
				return value.Dict(jwtStatusInvalid("bad_key", err.Error())), nil
			}
			_, header, payload, signingInput, signature, status, err := jwtParseForVerify(token)
			if err != nil {
				return value.Value{}, err
			}
			if status != nil {
				return value.Dict(status), nil
			}
			if alg, ok := header["alg"]; !ok || alg.Kind != value.KindString || alg.Str != "RS256" {
				return value.Dict(jwtStatusInvalid("alg_mismatch", "token alg must be RS256")), nil
			}
			h := sha256.Sum256([]byte(signingInput))
			if err := rsa.VerifyPKCS1v15(pub, crypto.SHA256, h[:], signature); err != nil {
				return value.Dict(jwtStatusInvalid("bad_signature", "invalid signature")), nil
			}
			if s := validateJWTClaims(payload, nowUnix); s != nil {
				return value.Dict(s), nil
			}
			return value.Dict(jwtStatusValid(header, payload)), nil
		},
	})
}

func registerJWTSignES256() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:           builtins.CryptoJWTSignES256,
			Name:         "__builtin_crypto_jwt_sign_es256",
			Arity:        3,
			ParamNames:   []string{"header", "payload", "privateKeyPem"},
			Params:       []builtins.TypeRef{{Kind: builtins.TypeDict, Elem: []builtins.TypeRef{{Kind: builtins.TypeAny}}}, {Kind: builtins.TypeDict, Elem: []builtins.TypeRef{{Kind: builtins.TypeAny}}}, {Kind: builtins.TypeBytes}},
			Result:       builtins.TypeRef{Kind: builtins.TypeString},
			ReceiverType: builtins.TypeVoid,
			MethodName:   "",
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			_, _, signingInput, err := jwtPrepareHeaderPayload(args, "ES256", "__builtin_crypto_jwt_sign_es256")
			if err != nil {
				return value.Value{}, err
			}
			pemBytes, err := requireBytesArg(args, 2, "__builtin_crypto_jwt_sign_es256")
			if err != nil {
				return value.Value{}, err
			}
			priv, err := parseECDSAPrivateKeyPEM(pemBytes)
			if err != nil {
				return value.Value{}, err
			}
			if priv.Curve != elliptic.P256() {
				return value.Value{}, fmt.Errorf("ES256 requires P-256 private key")
			}
			h := sha256.Sum256([]byte(signingInput))
			r, s, err := ecdsa.Sign(rand.Reader, priv, h[:])
			if err != nil {
				return value.Value{}, err
			}
			sig := ecdsaRawSignature(r, s, 32)
			return value.Str(signingInput + "." + jwtBase64.EncodeToString(sig)), nil
		},
	})
}

func registerJWTVerifyES256() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:           builtins.CryptoJWTVerifyES256,
			Name:         "__builtin_crypto_jwt_verify_es256",
			Arity:        3,
			ParamNames:   []string{"token", "publicKeyPem", "now"},
			Params:       []builtins.TypeRef{{Kind: builtins.TypeString}, {Kind: builtins.TypeBytes}, {Kind: builtins.TypeInt}},
			Result:       builtins.TypeRef{Kind: builtins.TypeDict, Elem: []builtins.TypeRef{{Kind: builtins.TypeAny}}},
			ReceiverType: builtins.TypeVoid,
			MethodName:   "",
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			token, err := requireStringArg(args, 0, "__builtin_crypto_jwt_verify_es256")
			if err != nil {
				return value.Value{}, err
			}
			pemBytes, err := requireBytesArg(args, 1, "__builtin_crypto_jwt_verify_es256")
			if err != nil {
				return value.Value{}, err
			}
			nowUnix, err := requireIntArg(args, 2, "__builtin_crypto_jwt_verify_es256")
			if err != nil {
				return value.Value{}, err
			}
			pub, err := parseECDSAPublicKeyPEM(pemBytes)
			if err != nil {
				return value.Dict(jwtStatusInvalid("bad_key", err.Error())), nil
			}
			if pub.Curve != elliptic.P256() {
				return value.Dict(jwtStatusInvalid("bad_key", "ES256 requires P-256 public key")), nil
			}
			_, header, payload, signingInput, signature, status, err := jwtParseForVerify(token)
			if err != nil {
				return value.Value{}, err
			}
			if status != nil {
				return value.Dict(status), nil
			}
			if alg, ok := header["alg"]; !ok || alg.Kind != value.KindString || alg.Str != "ES256" {
				return value.Dict(jwtStatusInvalid("alg_mismatch", "token alg must be ES256")), nil
			}
			r, s, ok := ecdsaRawSignatureParts(signature, 32)
			if !ok {
				return value.Dict(jwtStatusInvalid("bad_signature", "invalid signature format")), nil
			}
			h := sha256.Sum256([]byte(signingInput))
			if !ecdsa.Verify(pub, h[:], r, s) {
				return value.Dict(jwtStatusInvalid("bad_signature", "invalid signature")), nil
			}
			if s := validateJWTClaims(payload, nowUnix); s != nil {
				return value.Dict(s), nil
			}
			return value.Dict(jwtStatusValid(header, payload)), nil
		},
	})
}

func registerPasswordHash() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:           builtins.CryptoPasswordHash,
			Name:         "__builtin_crypto_password_hash",
			Arity:        1,
			ParamNames:   []string{"password"},
			Params:       []builtins.TypeRef{{Kind: builtins.TypeString}},
			Result:       builtins.TypeRef{Kind: builtins.TypeString},
			ReceiverType: builtins.TypeVoid,
			MethodName:   "",
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			password, err := requireStringArg(args, 0, "__builtin_crypto_password_hash")
			if err != nil {
				return value.Value{}, err
			}
			hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
			if err != nil {
				return value.Value{}, err
			}
			return value.Str(string(hash)), nil
		},
	})
}

func registerPasswordVerify() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:           builtins.CryptoPasswordVerify,
			Name:         "__builtin_crypto_password_verify",
			Arity:        2,
			ParamNames:   []string{"password", "encoded"},
			Params:       []builtins.TypeRef{{Kind: builtins.TypeString}, {Kind: builtins.TypeString}},
			Result:       builtins.TypeRef{Kind: builtins.TypeBool},
			ReceiverType: builtins.TypeVoid,
			MethodName:   "",
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			password, err := requireStringArg(args, 0, "__builtin_crypto_password_verify")
			if err != nil {
				return value.Value{}, err
			}
			encoded, err := requireStringArg(args, 1, "__builtin_crypto_password_verify")
			if err != nil {
				return value.Value{}, err
			}
			err = bcrypt.CompareHashAndPassword([]byte(encoded), []byte(password))
			if err == nil {
				return value.Bool(true), nil
			}
			if err == bcrypt.ErrMismatchedHashAndPassword || err == bcrypt.ErrHashTooShort {
				return value.Bool(false), nil
			}
			return value.Value{}, err
		},
	})
}

func requireBytesArg(args []interface{}, idx int, name string) ([]byte, error) {
	if len(args) <= idx {
		return nil, fmt.Errorf("%s expects %d arguments, got %d", name, idx+1, len(args))
	}
	v := args[idx].(value.Value)
	if v.Kind != value.KindBytes {
		return nil, fmt.Errorf("%s expects argument %d as bytes", name, idx+1)
	}
	return v.Bytes, nil
}

func requireStringArg(args []interface{}, idx int, name string) (string, error) {
	if len(args) <= idx {
		return "", fmt.Errorf("%s expects %d arguments, got %d", name, idx+1, len(args))
	}
	v := args[idx].(value.Value)
	if v.Kind != value.KindString {
		return "", fmt.Errorf("%s expects argument %d as string", name, idx+1)
	}
	return v.Str, nil
}

func requireIntArg(args []interface{}, idx int, name string) (int64, error) {
	if len(args) <= idx {
		return 0, fmt.Errorf("%s expects %d arguments, got %d", name, idx+1, len(args))
	}
	v := args[idx].(value.Value)
	if v.Kind != value.KindInt {
		return 0, fmt.Errorf("%s expects argument %d as int", name, idx+1)
	}
	return v.Int, nil
}

func requireDictArg(args []interface{}, idx int, name string) (map[string]value.Value, error) {
	if len(args) <= idx {
		return nil, fmt.Errorf("%s expects %d arguments, got %d", name, idx+1, len(args))
	}
	v := args[idx].(value.Value)
	if v.Kind != value.KindDict {
		return nil, fmt.Errorf("%s expects argument %d as dict<any>", name, idx+1)
	}
	return v.Dict, nil
}

func jwtPrepareHeaderPayload(args []interface{}, alg string, name string) (map[string]value.Value, map[string]value.Value, string, error) {
	header, err := requireDictArg(args, 0, name)
	if err != nil {
		return nil, nil, "", err
	}
	payload, err := requireDictArg(args, 1, name)
	if err != nil {
		return nil, nil, "", err
	}
	headerCopy := cloneDict(header)
	headerCopy["alg"] = value.Str(alg)
	headerCopy["typ"] = value.Str("JWT")
	headerJSON, err := encodeJWTObject(headerCopy)
	if err != nil {
		return nil, nil, "", err
	}
	payloadJSON, err := encodeJWTObject(payload)
	if err != nil {
		return nil, nil, "", err
	}
	signingInput := jwtBase64.EncodeToString(headerJSON) + "." + jwtBase64.EncodeToString(payloadJSON)
	return headerCopy, payload, signingInput, nil
}

func encodeJWTObject(dict map[string]value.Value) ([]byte, error) {
	obj := make(map[string]interface{}, len(dict))
	for k, v := range dict {
		j, err := valueToJSON(v)
		if err != nil {
			return nil, fmt.Errorf("unsupported JWT value for key %q: %w", k, err)
		}
		obj[k] = j
	}
	return json.Marshal(obj)
}

func valueToJSON(v value.Value) (interface{}, error) {
	switch v.Kind {
	case value.KindInt:
		return v.Int, nil
	case value.KindFloat:
		if math.IsNaN(v.Float) || math.IsInf(v.Float, 0) {
			return nil, fmt.Errorf("non-finite float")
		}
		return v.Float, nil
	case value.KindString:
		return v.Str, nil
	case value.KindBool:
		return v.Bool, nil
	case value.KindBytes:
		return jwtBase64.EncodeToString(v.Bytes), nil
	case value.KindOptional:
		if v.Optional == nil || !v.Optional.IsSome {
			return nil, nil
		}
		return valueToJSON(v.Optional.Value)
	case value.KindList:
		items := make([]interface{}, len(v.List))
		for i := range v.List {
			j, err := valueToJSON(v.List[i])
			if err != nil {
				return nil, err
			}
			items[i] = j
		}
		return items, nil
	case value.KindDict:
		obj := make(map[string]interface{}, len(v.Dict))
		for k, item := range v.Dict {
			j, err := valueToJSON(item)
			if err != nil {
				return nil, err
			}
			obj[k] = j
		}
		return obj, nil
	default:
		return nil, fmt.Errorf("kind %v", v.Kind)
	}
}

func jwtParseForVerify(token string) ([]string, map[string]value.Value, map[string]value.Value, string, []byte, map[string]value.Value, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, nil, nil, "", nil, jwtStatusInvalid("malformed", "token must have 3 parts"), nil
	}
	headerBytes, err := jwtBase64.DecodeString(parts[0])
	if err != nil {
		return nil, nil, nil, "", nil, jwtStatusInvalid("malformed", "invalid header encoding"), nil
	}
	payloadBytes, err := jwtBase64.DecodeString(parts[1])
	if err != nil {
		return nil, nil, nil, "", nil, jwtStatusInvalid("malformed", "invalid payload encoding"), nil
	}
	signature, err := jwtBase64.DecodeString(parts[2])
	if err != nil {
		return nil, nil, nil, "", nil, jwtStatusInvalid("malformed", "invalid signature encoding"), nil
	}
	header, err := parseJSONObjectToValueDict(headerBytes)
	if err != nil {
		return nil, nil, nil, "", nil, jwtStatusInvalid("malformed", "header is not a JSON object"), nil
	}
	payload, err := parseJSONObjectToValueDict(payloadBytes)
	if err != nil {
		return nil, nil, nil, "", nil, jwtStatusInvalid("malformed", "payload is not a JSON object"), nil
	}
	return parts, header, payload, parts[0] + "." + parts[1], signature, nil, nil
}

func parseJSONObjectToValueDict(data []byte) (map[string]value.Value, error) {
	dec := json.NewDecoder(strings.NewReader(string(data)))
	dec.UseNumber()
	var obj map[string]interface{}
	if err := dec.Decode(&obj); err != nil {
		return nil, err
	}
	out := make(map[string]value.Value, len(obj))
	for k, v := range obj {
		vv, err := jsonToValue(v)
		if err != nil {
			return nil, err
		}
		out[k] = vv
	}
	return out, nil
}

func jsonToValue(v interface{}) (value.Value, error) {
	switch val := v.(type) {
	case nil:
		return value.None(), nil
	case bool:
		return value.Bool(val), nil
	case string:
		return value.Str(val), nil
	case json.Number:
		raw := val.String()
		if !strings.ContainsAny(raw, ".eE") {
			i, err := val.Int64()
			if err == nil {
				return value.Int(i), nil
			}
		}
		f, err := val.Float64()
		if err != nil {
			return value.Value{}, err
		}
		return value.Float(f), nil
	case float64:
		return value.Float(val), nil
	case []interface{}:
		items := make([]value.Value, len(val))
		for i := range val {
			item, err := jsonToValue(val[i])
			if err != nil {
				return value.Value{}, err
			}
			items[i] = item
		}
		return value.List(items), nil
	case map[string]interface{}:
		d := make(map[string]value.Value, len(val))
		for k, item := range val {
			vv, err := jsonToValue(item)
			if err != nil {
				return value.Value{}, err
			}
			d[k] = vv
		}
		return value.Dict(d), nil
	default:
		return value.Value{}, fmt.Errorf("unsupported json value %T", v)
	}
}

func jwtStatusValid(header map[string]value.Value, payload map[string]value.Value) map[string]value.Value {
	return map[string]value.Value{
		"valid":   value.Bool(true),
		"header":  value.Dict(header),
		"payload": value.Dict(payload),
	}
}

func jwtStatusInvalid(reason string, message string) map[string]value.Value {
	return map[string]value.Value{
		"valid":  value.Bool(false),
		"reason": value.Str(reason),
		"error":  value.Str(message),
	}
}

func validateJWTClaims(payload map[string]value.Value, nowUnix int64) map[string]value.Value {
	if exp, ok := payload["exp"]; ok {
		expVal, ok := claimNumeric(exp)
		if !ok {
			return jwtStatusInvalid("invalid_claim", "exp must be numeric")
		}
		if nowUnix >= expVal {
			return jwtStatusInvalid("expired", "token is expired")
		}
	}
	if nbf, ok := payload["nbf"]; ok {
		nbfVal, ok := claimNumeric(nbf)
		if !ok {
			return jwtStatusInvalid("invalid_claim", "nbf must be numeric")
		}
		if nowUnix < nbfVal {
			return jwtStatusInvalid("not_before", "token is not active yet")
		}
	}
	if iat, ok := payload["iat"]; ok {
		iatVal, ok := claimNumeric(iat)
		if !ok {
			return jwtStatusInvalid("invalid_claim", "iat must be numeric")
		}
		if iatVal > nowUnix+300 {
			return jwtStatusInvalid("invalid_iat", "iat is in the future")
		}
	}
	return nil
}

func claimNumeric(v value.Value) (int64, bool) {
	switch v.Kind {
	case value.KindInt:
		return v.Int, true
	case value.KindFloat:
		return int64(v.Float), true
	default:
		return 0, false
	}
}

func parseRSAPrivateKeyPEM(pemBytes []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, fmt.Errorf("invalid PEM private key")
	}
	if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return key, nil
	}
	pk, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("invalid RSA private key")
	}
	key, ok := pk.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("private key is not RSA")
	}
	return key, nil
}

func parseRSAPublicKeyPEM(pemBytes []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, fmt.Errorf("invalid PEM public key")
	}
	if key, err := x509.ParsePKCS1PublicKey(block.Bytes); err == nil {
		return key, nil
	}
	pk, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("invalid RSA public key")
	}
	key, ok := pk.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("public key is not RSA")
	}
	return key, nil
}

func parseECDSAPrivateKeyPEM(pemBytes []byte) (*ecdsa.PrivateKey, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, fmt.Errorf("invalid PEM private key")
	}
	if key, err := x509.ParseECPrivateKey(block.Bytes); err == nil {
		return key, nil
	}
	pk, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("invalid ECDSA private key")
	}
	key, ok := pk.(*ecdsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("private key is not ECDSA")
	}
	return key, nil
}

func parseECDSAPublicKeyPEM(pemBytes []byte) (*ecdsa.PublicKey, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, fmt.Errorf("invalid PEM public key")
	}
	pk, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("invalid ECDSA public key")
	}
	key, ok := pk.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("public key is not ECDSA")
	}
	return key, nil
}

func ecdsaRawSignature(r *big.Int, s *big.Int, size int) []byte {
	out := make([]byte, size*2)
	rb := r.Bytes()
	sb := s.Bytes()
	copy(out[size-len(rb):size], rb)
	copy(out[2*size-len(sb):], sb)
	return out
}

func ecdsaRawSignatureParts(sig []byte, size int) (*big.Int, *big.Int, bool) {
	if len(sig) != size*2 {
		return nil, nil, false
	}
	r := new(big.Int).SetBytes(sig[:size])
	s := new(big.Int).SetBytes(sig[size:])
	if r.Sign() <= 0 || s.Sign() <= 0 {
		return nil, nil, false
	}
	return r, s, true
}

func registerRandomBytes() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:           builtins.CryptoRandomBytes,
			Name:         "__builtin_crypto_random_bytes",
			Arity:        1,
			ParamNames:   []string{"n"},
			Params:       []builtins.TypeRef{{Kind: builtins.TypeInt}},
			Result:       builtins.TypeRef{Kind: builtins.TypeBytes},
			ReceiverType: builtins.TypeVoid,
			MethodName:   "",
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			n, err := requireIntArg(args, 0, "__builtin_crypto_random_bytes")
			if err != nil {
				return value.Value{}, err
			}
			if n <= 0 || n > 1024 {
				return value.Value{}, fmt.Errorf("__builtin_crypto_random_bytes: n must be between 1 and 1024, got %d", n)
			}
			buf := make([]byte, n)
			if _, err := rand.Read(buf); err != nil {
				return value.Value{}, fmt.Errorf("__builtin_crypto_random_bytes: %w", err)
			}
			return value.Bytes(buf), nil
		},
	})
}

func cloneDict(src map[string]value.Value) map[string]value.Value {
	dst := make(map[string]value.Value, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

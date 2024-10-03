// package jwt implements the JSON Web Token (JWT) standard as per rfc7519
package jwt

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"golang.org/x/exp/maps"
)

type headerJOSE struct {
	Typ string `json:"typ"`
	Alg string `json:"alg"`
}

var supportedAlgorithms = map[string]struct{}{
	"HS256": {},
	// "HS384": {},
	// "HS512": {},
}

// Encode generates a JWT token with the given claims, secret and algorithm.
func Encode(claims map[string]any, secret, algorithm string) (string, error) {
	if _, ok := supportedAlgorithms[algorithm]; !ok {
		return "", fmt.Errorf("unsupported algorithm %s", algorithm)
	}
	header := headerJOSE{
		Typ: "JWT",
		Alg: algorithm,
	}
	buf, err := json.Marshal(&header)
	if err != nil {
		return "", err
	}
	h := base64.RawURLEncoding.EncodeToString(buf)
	keys := maps.Keys(claims)
	sort.Strings(keys)
	m := make(map[string]any)
	for _, k := range keys {
		m[k] = claims[k]
	}
	buf, err = json.Marshal(m)
	if err != nil {
		return "", err
	}
	c := base64.RawURLEncoding.EncodeToString(buf)
	signed := signJWT(secret, h, c)
	signature := base64.RawURLEncoding.EncodeToString(signed)
	// concat each encoded part with a period '.' separator
	return h + "." + c + "." + signature, nil
}

// https://datatracker.ietf.org/doc/html/rfc7519#section-7.2
func Validate(token string, secret string) bool {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return false
	}
	header, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return false
	}
	var h headerJOSE
	if err := json.Unmarshal(header, &h); err != nil {
		return false
	}
	if h.Typ != "JWT" {
		return false
	}
	if _, ok := supportedAlgorithms[h.Alg]; !ok {
		return false
	}
	// verify signature by string comparison
	verify := signJWT(secret, base64.RawURLEncoding.EncodeToString(header), parts[1], parts[2])
	verified := base64.RawURLEncoding.EncodeToString(verify)
	return verified == parts[2]
}

func signJWT(key string, parts ...string) []byte {
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(parts[0] + "." + parts[1]))
	return h.Sum(nil)
}

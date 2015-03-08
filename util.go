package relayR

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"unicode"
	"unicode/utf8"
)

func jsonResponse(w http.ResponseWriter) {
	w.Header().Set("Content-type", "application/json")
}

func generateConnectionID() string {
	rb := make([]byte, 32)
	rand.Read(rb)
	rs := base64.URLEncoding.EncodeToString(rb)
	return rs
}

func lowerFirst(s string) string {
	if s == "" {
		return ""
	}
	r, n := utf8.DecodeRuneInString(s)
	return string(unicode.ToLower(r)) + s[n:]
}

func upperFirst(s string) string {
	if s == "" {
		return ""
	}
	r, n := utf8.DecodeRuneInString(s)
	return string(unicode.ToUpper(r)) + s[n:]
}

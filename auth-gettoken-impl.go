package ginmiddleware

import (
	"errors"
	"net/http"
	"strings"
)

var (
	ErrorHeaderAuthFormat = errors.New("error header auth format")
)

var (
	JWTHeaderAuthToken   = HeaderAuthToken{Key: "Authorization", ValueHeader: "bearer"}
	BasicHeaderAuthToken = HeaderAuthToken{Key: "Authorization", ValueHeader: "basic"}
	JWTQueryAuthToken    = QueryAuthToken{Key: "access_token"}
)

type customGetAuthToken struct {
	fun func(*http.Request) (string, error)
}

func (custom *customGetAuthToken) GetAuthToken(req *http.Request) (string, error) {
	return custom.fun(req)
}

func NewCustomGetAuthToken(fun func(*http.Request) (string, error)) IgetAuthToken {
	return &customGetAuthToken{
		fun: fun,
	}
}

type HeaderAuthToken struct {
	Key         string
	ValueHeader string
}

func (h *HeaderAuthToken) GetAuthToken(r *http.Request) (string, error) {
	authHeader := r.Header.Get(h.Key)
	if authHeader == "" {
		return "", nil // No error, just no token
	}
	if h.ValueHeader == "" {
		return authHeader, nil
	}

	authHeaderParts := strings.Split(authHeader, " ")
	if len(authHeaderParts) != 2 || !strings.EqualFold(authHeaderParts[0], h.ValueHeader) {
		return "", ErrorHeaderAuthFormat
	}

	return authHeaderParts[1], nil
}

type QueryAuthToken struct {
	Key string
}

func (q *QueryAuthToken) GetAuthToken(r *http.Request) (string, error) {
	token := r.URL.Query().Get(q.Key)
	return token, nil
}

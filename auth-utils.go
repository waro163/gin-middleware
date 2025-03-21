package ginmiddleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type IgetAuthToken interface {
	GetAuthToken(*http.Request) (string, error)
}

var (
	_ IgetAuthToken = (*HeaderAuthToken)(nil)
	_ IgetAuthToken = (*QueryAuthToken)(nil)
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

var (
	ErrorHeaderAuthFormat = errors.New("error header auth format")
)

func (h *HeaderAuthToken) GetAuthToken(r *http.Request) (string, error) {
	authHeader := r.Header.Get(h.Key)
	if authHeader == "" {
		return "", nil // No error, just no token
	}
	if h.ValueHeader == "" {
		return authHeader, nil
	}

	authHeaderParts := strings.Split(authHeader, " ")
	if len(authHeaderParts) != 2 || strings.ToLower(authHeaderParts[0]) != h.ValueHeader {
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

type IAuthenticator interface {
	Authenticate(token string) (interface{}, *Error)
}

var (
	_ IAuthenticator = (*JWTAuthenticator)(nil)
)

type JWTAuthenticator struct {
	Audience        string
	Issuer          string
	GetPublicSecret func(*jwt.Token) (interface{}, error)
}

func (a *JWTAuthenticator) Authenticate(token string) (interface{}, *Error) {
	if token == "" {
		return nil, &Error{
			Code:    -1,
			Message: "missing authorization token",
		}
	}
	jwtToken, err := jwt.Parse(token, a.GetPublicSecret, jwt.WithIssuer(a.Issuer), jwt.WithAudience(a.Audience))
	if err != nil {
		return nil, &Error{
			Code:    -2,
			Message: err.Error(),
		}
	}
	if !jwtToken.Valid {
		return nil, &Error{
			Code:    -3,
			Message: "invalid jwt token",
		}
	}
	claims, ok := jwtToken.Claims.(jwt.MapClaims)
	if !ok {
		return nil, &Error{
			Code:    -4,
			Message: "invalid jwt token",
		}
	}

	return map[string]interface{}(claims), nil
}

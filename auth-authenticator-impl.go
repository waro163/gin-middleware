package ginmiddleware

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/waro163/gin-middleware/utils"
)

type CommonJwtAuthenticator struct {
	Options  []jwt.ParserOption
	JwtParse utils.IJwtParser
}

func (auth *CommonJwtAuthenticator) Authenticate(token string) (interface{}, *Error) {
	if token == "" {
		return nil, &Error{
			Code:    -1,
			Message: "missing authorization token",
		}
	}
	claims := jwt.MapClaims{}
	_, err := auth.JwtParse.ParseJwtToken(token, &claims, auth.Options...)
	if err != nil {
		return nil, &Error{
			Code:    -2,
			Message: err.Error(),
		}
	}
	return claims, nil
}

type CustomJwtAuthenticator struct {
	Audience string
	Issuer   string
	Subject  string
	JwtParse utils.IJwtParser
}

type MyCustomClaims struct {
	Uid string `json:"uid"`
	jwt.RegisteredClaims
}

func (auth *CustomJwtAuthenticator) Authenticate(token string) (interface{}, *Error) {
	if token == "" {
		return nil, &Error{
			Code:    -1,
			Message: "missing authorization token",
		}
	}
	claims := MyCustomClaims{}
	_, err := auth.JwtParse.ParseJwtToken(token, &claims, jwt.WithAudience(auth.Audience), jwt.WithIssuer(auth.Issuer), jwt.WithIssuedAt(), jwt.WithExpirationRequired(), jwt.WithSubject(auth.Subject))
	if err != nil {
		return nil, &Error{
			Code:    -2,
			Message: err.Error(),
		}
	}
	return claims, nil
}

type M2mJwtAuthenticator struct {
	Audience string
	Issuer   string
	Scope    string
	JwtParse utils.IJwtParser
}

type M2mClaims struct {
	jwt.RegisteredClaims
	Scp string `json:"scp,omitempty"`
}

func (auth *M2mJwtAuthenticator) Authenticate(token string) (interface{}, *Error) {
	if token == "" {
		return nil, &Error{
			Code:    -1,
			Message: "missing authorization token",
		}
	}
	parseToken, err := auth.JwtParse.ParseJwtToken(token, &M2mClaims{}, jwt.WithIssuer(auth.Issuer), jwt.WithIssuedAt(), jwt.WithAudience(auth.Audience), jwt.WithExpirationRequired())
	if err != nil {
		return nil, &Error{
			Code:    -2,
			Message: err.Error(),
		}
	}
	claims, ok := parseToken.Claims.(*M2mClaims)
	if !ok {
		return nil, &Error{
			Code:    -3,
			Message: "invalid claims",
		}
	}
	if auth.Scope != "" && claims.Scp != auth.Scope {
		return nil, &Error{
			Code:    -4,
			Message: "invalid scp",
		}
	}
	return claims, nil
}

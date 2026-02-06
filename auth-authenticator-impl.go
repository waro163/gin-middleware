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

func NewCustomJwtAuthenticator(audience, issuer, subject string, jwksUrl string) *CustomJwtAuthenticator {
	jwtParser := utils.NewJwtParse(jwksUrl, nil, nil)
	return &CustomJwtAuthenticator{
		Audience: audience,
		Issuer:   issuer,
		Subject:  subject,
		JwtParse: jwtParser,
	}
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
	Subject  string
	Scope    string
	JwtParse utils.IJwtParser
}

func NewM2mJwtAuthenticator(audience, issuer, subject, scope, jwksUrl string) *M2mJwtAuthenticator {
	jwtParser := utils.NewJwtParse(jwksUrl, nil, nil)
	return &M2mJwtAuthenticator{
		Audience: audience,
		Issuer:   issuer,
		Subject:  subject,
		Scope:    scope,
		JwtParse: jwtParser,
	}
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
	claims := M2mClaims{}
	_, err := auth.JwtParse.ParseJwtToken(token, &claims, jwt.WithIssuer(auth.Issuer), jwt.WithIssuedAt(), jwt.WithAudience(auth.Audience), jwt.WithExpirationRequired(), jwt.WithSubject(auth.Subject))
	if err != nil {
		return nil, &Error{
			Code:    -2,
			Message: err.Error(),
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

package ginmiddleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	AuthPayload = "auth-payload"

	ACCESS_TOKEN_SUBJECT  = "access_token"
	REFRESH_TOKEN_SUBJECT = "refresh_token"
)

type IgetAuthToken interface {
	GetAuthToken(*http.Request) (string, error)
}

var (
	_ IgetAuthToken = (*HeaderAuthToken)(nil)
	_ IgetAuthToken = (*QueryAuthToken)(nil)
)

type IAuthenticator interface {
	Authenticate(token string) (interface{}, *Error)
}

var (
	_ IAuthenticator = (*CommonJwtAuthenticator)(nil)
	_ IAuthenticator = (*CustomJwtAuthenticator)(nil)
	_ IAuthenticator = (*M2mJwtAuthenticator)(nil)
)

func NewAuthMiddleware(getToken IgetAuthToken, authenticator IAuthenticator) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := getToken.GetAuthToken(c.Request)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, &Error{
				Code:    http.StatusUnauthorized,
				Message: err.Error(),
			})
			return
		}
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, &Error{
				Code:    http.StatusUnauthorized,
				Message: "not found auth token",
			})
			return
		}
		claims, mErr := authenticator.Authenticate(token)
		if mErr != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, &Error{
				Code:    http.StatusUnauthorized,
				Message: mErr.Message,
			})
			return
		}
		c.Set(AuthPayload, claims)
		c.Next()
	}
}

func GetAuthPayload(c *gin.Context) (interface{}, bool) {
	return c.Get(AuthPayload)
}

type AuthMiddleware struct {
	GetToken      IgetAuthToken
	Authenticator IAuthenticator
}

func (auth *AuthMiddleware) ValidateAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := auth.GetToken.GetAuthToken(c.Request)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, &Error{
				Code:    http.StatusUnauthorized,
				Message: err.Error(),
			})
			return
		}
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, &Error{
				Code:    http.StatusUnauthorized,
				Message: "not found auth token",
			})
			return
		}
		claims, mErr := auth.Authenticator.Authenticate(token)
		if mErr != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, &Error{
				Code:    http.StatusUnauthorized,
				Message: mErr.Message,
			})
			return
		}
		c.Set(AuthPayload, claims)
		c.Next()
	}
}

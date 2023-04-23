package ginmiddleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

var AuthPayload string = "auth-payload"

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

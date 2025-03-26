package utils

import (
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
)

type PublicKey struct {
	Kid    string `json:"kid"`
	Public string `json:"public"`
}

type IJwtParser interface {
	ParseJwtToken(string, jwt.Claims, ...jwt.ParserOption) (*jwt.Token, error)
}

type IRequestClient interface {
	Do(*http.Request) (*http.Response, error)
}

var DefaultClient IRequestClient = http.DefaultClient

type JwtParse struct {
	jwksUrl string
	client  IRequestClient
	cache   ICache
}

func NewJwtParse(jwksUrl string, client IRequestClient, cache ICache) IJwtParser {
	parse := &JwtParse{
		jwksUrl: jwksUrl,
		client:  client,
		cache:   cache,
	}
	if client == nil {
		parse.client = DefaultClient
	}
	if cache == nil {
		parse.cache = DefaultMemoryCache
	}
	return parse
}

func (p *JwtParse) ParseJwtToken(token string, claims jwt.Claims, options ...jwt.ParserOption) (*jwt.Token, error) {
	parseToken, err := jwt.ParseWithClaims(token, claims, p.getPublicSecret, options...)
	return parseToken, err
}

func (p *JwtParse) getPublicSecret(token *jwt.Token) (interface{}, error) {
	if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
		return nil, fmt.Errorf("unexpected signed method: %s", token.Header["alg"])
	}
	kid, ok := token.Header["kid"].(string)
	if !ok || kid == "" {
		return nil, fmt.Errorf("invalid kid %s", kid)
	}
	publicKey, err := p.cache.Get(kid)
	if err == nil && publicKey != nil {
		return publicKey, nil
	}
	req, err := http.NewRequest(http.MethodGet, p.jwksUrl, nil)
	if err != nil {
		return nil, err
	}
	response, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	var pubCerts []PublicKey
	if err := json.NewDecoder(response.Body).Decode(&pubCerts); err != nil {
		return nil, err
	}
	for _, pubCert := range pubCerts {
		rsaPublicKey, err := convertPublicCert(pubCert.Public)
		if err != nil {
			continue
		}
		p.cache.Set(pubCert.Kid, rsaPublicKey, 0)
	}
	publicKey, err = p.cache.Get(kid)
	if err == nil && publicKey != nil {
		return publicKey, nil
	}
	return nil, fmt.Errorf("not found kid %s in jwks", kid)
}

func convertPublicCert(cert string) (interface{}, error) {
	rsaPublic, err := base64.StdEncoding.DecodeString(cert)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(rsaPublic)
	rsaPublicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return rsaPublicKey, nil
}

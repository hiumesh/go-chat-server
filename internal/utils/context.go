package utils

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

type AMREntry struct {
	Method    string `json:"method"`
	Timestamp int64  `json:"timestamp"`
	Provider  string `json:"provider,omitempty"`
}

type AccessTokenClaims struct {
	jwt.StandardClaims
	Email                         string                 `json:"email"`
	Phone                         string                 `json:"phone"`
	AppMetaData                   map[string]interface{} `json:"app_metadata"`
	UserMetaData                  map[string]interface{} `json:"user_metadata"`
	Role                          string                 `json:"role"`
	AuthenticatorAssuranceLevel   string                 `json:"aal,omitempty"`
	AuthenticationMethodReference []AMREntry             `json:"amr,omitempty"`
	SessionId                     string                 `json:"session_id,omitempty"`
}

type contextKey string

func (c contextKey) String() string {
	return "gotrue api context key " + string(c)
}

const (
	tokenKey     = contextKey("jwt")
	requestIDKey = contextKey("request_id")
)

// withToken adds the JWT token to the context.
func WithToken(ctx *gin.Context, token *jwt.Token) {
	ctx.Set(string(tokenKey), token)
}

// getToken reads the JWT token from the context.
func GetToken(ctx *gin.Context) *jwt.Token {
	obj, exists := ctx.Get(string(tokenKey))
	if !exists || obj == nil {
		return nil
	}

	return obj.(*jwt.Token)
}

func GetClaims(ctx *gin.Context) *AccessTokenClaims {
	token := GetToken(ctx)
	if token == nil {
		return nil
	}
	return token.Claims.(*AccessTokenClaims)
}

func WithRequestID(ctx *gin.Context, id string) {
	ctx.Set(string(requestIDKey), id)
}

// getRequestID reads the request ID from the context.
func GetRequestID(ctx *gin.Context) string {
	obj, exists := ctx.Get(string(requestIDKey))
	if !exists || obj == nil {
		return ""
	}

	return obj.(string)
}

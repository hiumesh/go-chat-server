package api

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/hiumesh/go-chat-server/internal/utils"
	"github.com/sirupsen/logrus"
)

func (a *API) extractBearerToken(ctx *gin.Context) (string, *utils.HTTPError) {
	authHeader := ctx.Request.Header.Get("Authorization")
	matches := bearerRegexp.FindStringSubmatch(authHeader)
	if len(matches) != 2 {
		return "", utils.UnauthorizedError("This endpoint requires a Bearer token")
	}

	return matches[1], nil
}

func (a *API) parseJWTClaims(bearer string, ctx *gin.Context) (context.Context, *utils.HTTPError) {
	config := a.config

	p := jwt.Parser{ValidMethods: []string{jwt.SigningMethodHS256.Name}}
	token, err := p.ParseWithClaims(bearer, &utils.AccessTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.JWT.Secret), nil
	})
	if err != nil {
		return ctx, utils.UnauthorizedError("invalid JWT: unable to parse or verify signature, %v", err)
	}
	utils.WithToken(ctx, token)
	return ctx, nil
}

// requireAuthentication checks incoming requests for tokens presented using the Authorization header
func (a *API) requireAuthentication(ctx *gin.Context) {
	token, err := a.extractBearerToken(ctx)
	config := a.config
	if err != nil {
		a.clearCookieTokens(config, ctx.Writer)
		logrus.Error("Authentication Error: ", err)
		ctx.AbortWithStatusJSON(err.Code, err)
		return
	}

	_, err = a.parseJWTClaims(token, ctx)
	if err != nil {
		a.clearCookieTokens(config, ctx.Writer)
		logrus.Error("Authentication Error: ", err)
		ctx.AbortWithStatusJSON(err.Code, err)
		return
	}

}

package middleware

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/nhassl3/simplebank/internals/lib/logger/sl"
	"github.com/nhassl3/simplebank/internals/lib/token"
)

const (
	authorizationHeaderKey            = "authorization"
	authorizationTypeBearer           = "bearer"
	AuthorizationPayloadKey           = "authorization_payload"
	ErrorAuthorizationHeader          = "authorization header is not provided"
	ErrorInvalidAuthorizationHeader   = "invalid authorization header format"
	ErrorUnsupportedAuthorizationType = "unsupported authorization type"
	ErrorTokenIsExpired               = "token is expired"
	ErrorInvalidToken                 = "token is invalid"
	ErrorFailedToVerify               = "failed to verify token"
	ErrorUnauthorized                 = "unauthorized"
)

// AuthMiddleware creates a gin middleware for authorization
func AuthMiddleware(tokenMaker token.Maker, log *slog.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authorizationHeader := ctx.GetHeader(authorizationHeaderKey)

		if len(authorizationHeader) == 0 {
			log.Warn(ErrorAuthorizationHeader)
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": ErrorAuthorizationHeader})
			return
		}

		fields := strings.Fields(authorizationHeader)
		if len(fields) < 2 {
			log.Warn(ErrorInvalidAuthorizationHeader)
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": ErrorInvalidAuthorizationHeader})
			return
		}

		authorizationType := strings.ToLower(fields[0])
		if authorizationType != authorizationTypeBearer {
			log.Warn(ErrorUnsupportedAuthorizationType, slog.String("type", authorizationType))
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": ErrorUnsupportedAuthorizationType})
			return
		}

		accessToken := fields[1]
		payload, err := tokenMaker.VerifyToken(accessToken)
		if err != nil {
			if errors.Is(err, token.ErrExpiredToken) {
				log.Warn(ErrorTokenIsExpired, sl.Err(err))
				ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": ErrorTokenIsExpired})
				return
			} else if errors.Is(err, token.ErrInvalidToken) {
				log.Warn(ErrorInvalidToken, sl.Err(err))
				if accessToken == "undefined" {
					ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": ErrorUnauthorized})
					return
				}
				ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": ErrorInvalidToken})
				return
			}
			log.Error(ErrorFailedToVerify, sl.Err(err))
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": ErrorFailedToVerify})
			return
		}

		ctx.Set(AuthorizationPayloadKey, payload)
		ctx.Next()
	}
}

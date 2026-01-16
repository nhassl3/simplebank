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
	authorizationHeaderKey  = "authorization"
	authorizationTypeBearer = "bearer"
	authorizationPayloadKey = "authorization_payload"
)

// AuthMiddleware creates a gin middleware for authorization
func AuthMiddleware(tokenMaker token.Maker, log *slog.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		log = log.With("op", "middleware.AuthMiddleware")

		authorizationHeader := ctx.GetHeader(authorizationHeaderKey)

		if len(authorizationHeader) == 0 {
			log.Warn("authorization header is not provided")
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header is not provided"})
			ctx.Abort()
			return
		}

		fields := strings.Fields(authorizationHeader)
		if len(fields) < 2 {
			log.Warn("invalid authorization header format")
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			ctx.Abort()
			return
		}

		authorizationType := strings.ToLower(fields[0])
		if authorizationType != authorizationTypeBearer {
			log.Warn("unsupported authorization type", slog.String("type", authorizationType))
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unsupported authorization type"})
			ctx.Abort()
			return
		}

		accessToken := fields[1]
		payload, err := tokenMaker.VerifyToken(accessToken)
		if err != nil {
			if errors.Is(err, token.ErrExpiredToken) {
				log.Warn("token is expired", sl.Err(err))
				ctx.JSON(http.StatusUnauthorized, gin.H{"error": "token is expired"})
				ctx.Abort()
				return
			}
			if errors.Is(err, token.ErrInvalidToken) {
				log.Warn("token is invalid", sl.Err(err))
				ctx.JSON(http.StatusUnauthorized, gin.H{"error": "token is invalid"})
				ctx.Abort()
				return
			}
			log.Error("failed to verify token", sl.Err(err))
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to verify token"})
			ctx.Abort()
			return
		}

		ctx.Set(authorizationPayloadKey, payload)
		ctx.Next()
	}
}

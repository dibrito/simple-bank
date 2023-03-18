package api

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/dibrito/simple-bank/token"
	"github.com/gin-gonic/gin"
)

const (
	authHeaderKey           = "authorization"
	authorizationTypeBearer = "bearer"
	authPayloadKey          = "authorization_payload"
)

func authMiddleware(tokenMaker token.Maker) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// extract header
		header := ctx.GetHeader(authHeaderKey)
		if len(header) == 0 {
			err := errors.New("authorization header is not provided")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}

		// broke header
		fields := strings.Fields(header)
		if len(fields) != 2 {
			err := errors.New("invalid authorization header format")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}

		// check auth type
		inAuthType := strings.ToLower(fields[0])
		if inAuthType != authorizationTypeBearer {
			err := fmt.Errorf("unsuported authorization type: %v", inAuthType)
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}

		payload, err := tokenMaker.VerifyToken(fields[1])
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}

		ctx.Set(authPayloadKey, payload)
		ctx.Next()
	}
}

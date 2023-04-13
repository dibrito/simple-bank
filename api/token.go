package api

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type renewAccessTokenRequest struct {
	// alphanum ASCII alpha nums only
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type renewAccessTokenResponse struct {
	AccessToken         string    `json:"access_token"`
	AccessTokenExpireAt time.Time `json:"access_token_expire_at"`
}

func (s *Server) renewAccessToken(ctx *gin.Context) {
	var req renewAccessTokenRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	refreshTokenPayload, err := s.tokenMaker.VerifyToken(req.RefreshToken)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	// retrieve session

	session, err := s.store.GetSession(ctx, refreshTokenPayload.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	if session.IsBlocked {
		ctx.JSON(http.StatusUnauthorized, errorResponse(fmt.Errorf("blocked session")))
		return
	}

	if session.Username != refreshTokenPayload.Username {
		ctx.JSON(http.StatusUnauthorized, errorResponse(fmt.Errorf("incorrect user session")))
		return
	}

	if session.RefreshToken != req.RefreshToken {
		ctx.JSON(http.StatusUnauthorized, errorResponse(fmt.Errorf("mismatched session token")))
		return
	}

	if time.Now().After(session.ExpiresAt) {
		ctx.JSON(http.StatusUnauthorized, errorResponse(fmt.Errorf("expired session")))
		return
	}

	// create new token
	accessToken, accessTokenPayload, err := s.tokenMaker.CreateToken(
		refreshTokenPayload.Username,
		s.config.AccessDuration)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	resp := renewAccessTokenResponse{
		AccessToken:         accessToken,
		AccessTokenExpireAt: accessTokenPayload.ExpireAt,
	}

	ctx.JSON(http.StatusOK, resp)
}

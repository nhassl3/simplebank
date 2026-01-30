package tests

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nhassl3/simplebank/internals/http/middleware"
	"github.com/nhassl3/simplebank/internals/lib/token"
	"github.com/nhassl3/simplebank/tests/suit"
	"github.com/stretchr/testify/require"
)

func addAuthorization(
	t *testing.T,
	request *http.Request,
	tokenMaker token.Maker,
	authType,
	username string,
) {
	tokenString, err := tokenMaker.CreateToken(username, nil)
	require.NoError(t, err)
	require.NotEmpty(t, tokenString)

	request.Header.Set("Authorization", fmt.Sprintf("%s %s", authType, tokenString))
}

func TestAuthMiddleware(t *testing.T) {
	testCases := []struct {
		Name          string
		SetupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		CheckResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			Name: "OK",
			SetupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, "Bearer", "hach228")
			},
			CheckResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			Name: "NoAuthorization",
			SetupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				//addAuthorization(t, request, tokenMaker, "Bearer", "hach228")
			},
			CheckResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			Name: "UnsupportedAuthorization",
			SetupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, "Bearer1", "hach228")
			},
			CheckResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			Name: "InvalidAuthorizationFormat",
			SetupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, "", "hach228")
			},
			CheckResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			Name: "ExpiredToken",
			SetupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, "Bearer", "hach228")
			},
			CheckResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
	}

	log := slog.Default()
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			var (
				s   *suit.Suite
				err error
			)
			if tc.Name == "ExpiredToken" {
				s, err = suit.NewSuite(t, -1*time.Minute)
				require.NoError(t, err)
				require.NotEmpty(t, s)
			} else {
				s, err = suit.NewSuite(t, 1*time.Minute)
				require.NoError(t, err)
				require.NotEmpty(t, s)
			}

			s.Server.Router.GET("/api/auth", middleware.AuthMiddleware(s.TGPMaker, log), func(ctx *gin.Context) {
				ctx.JSON(http.StatusOK, gin.H{})
			})

			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, "/api/auth", nil)

			tc.SetupAuth(t, request, s.TGPMaker)
			s.Server.Router.ServeHTTP(recorder, request)
			tc.CheckResponse(t, recorder)

			s.Finish()
		})
	}
}

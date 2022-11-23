package auth

import (
	"context"
	"net/http"
	"os"

	"google-sheets-api/middleware"

	"github.com/gin-gonic/gin"

	"golang.org/x/oauth2"
)

func googleOAuth(c *gin.Context) {
	config, _ := c.MustGet("config").(*oauth2.Config)
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	c.String(http.StatusOK, authURL)
}

func redirect(c *gin.Context) {
	authCode := c.Query("code")
	config, _ := c.MustGet("config").(*oauth2.Config)
	tok, err := config.Exchange(context.TODO(), authCode)
	if err == nil {
		middleware.SaveToken(middleware.TokFile, tok)
	}
}

func removeTokenFile(c *gin.Context) {
	err := os.Remove(middleware.TokFile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	c.Writer.WriteHeader(200)
}

func InitAuthHandler(e *gin.Engine) {
	e.Use(middleware.ConfigMiddleware)
	{
		e.GET("/login", googleOAuth)
		e.GET("/logout", removeTokenFile)
		e.GET("/redirect", redirect)
	}
}

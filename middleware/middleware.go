package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

const TokFile = "token.json"

func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

func SaveToken(path string, token *oauth2.Token) {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func ConfigMiddleware(c *gin.Context) {
	b, err := os.ReadFile("credentials.json")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot read credentals.json"})
		c.Abort()
	}
	config, err := google.ConfigFromJSON(b, "https://www.googleapis.com/auth/spreadsheets")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		c.Abort()
	} else {
		c.Set("config", config)
	}
}

func ServiceMiddleware(c *gin.Context) {
	ctx := context.TODO()
	config, _ := c.MustGet("config").(*oauth2.Config)
	tok, err := tokenFromFile(TokFile)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	expired := tok.Expiry.Before(time.Now())
	if expired {
		newToken, err := config.TokenSource(context.TODO(), tok).Token()
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			c.Abort()
			return
		}
		if newToken.AccessToken != tok.AccessToken {
			SaveToken(TokFile, newToken)
		}
	}

	client := config.Client(context.Background(), tok)

	srv, err := sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		c.Abort()
	} else {
		c.Set("service", srv)
	}
}

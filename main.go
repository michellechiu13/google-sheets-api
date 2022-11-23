package main

import (
	"flag"

	auth "google-sheets-api/auth"
	sheet "google-sheets-api/sheets"

	"github.com/gin-gonic/gin"
)

func main() {
	debug := flag.Bool("debug", false, "Debug mode")
	flag.Parse()

	if !*debug {
		gin.SetMode(gin.ReleaseMode)
	}

	e := gin.New()
	auth.InitAuthHandler(e)
	sheet.InitSheetHandler(e)
	e.Run()
}

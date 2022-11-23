package sheet

import (
	"google-sheets-api/middleware"

	"github.com/gin-gonic/gin"
)

type RawData struct {
	Values [][]interface{}
}

func InitSheetHandler(e *gin.Engine) {
	g := e.Group("sheet/:id")
	g.Use(middleware.ServiceMiddleware)
	{
		g.GET("/", returnAllSheets)
		g.GET(":sheetId", returnJsonSheetData)
		g.GET(":sheetId/keys", returnJsonSheetKeys)
		g.GET(":sheetId/id/:itemId", returnSingleItem)
		g.GET(":sheetId/search", returnSearchData)
		g.PUT(":sheetId/id/:itemId", updateSingleItem)
		g.POST(":sheetId", appendSingleItem)
		g.DELETE(":sheetId/id/:itemId", deleteSingleItem)
	}

	raw := g.Group("raw")
	{
		raw.GET(":range", returnSheetData)
		raw.PUT(":range", updateSheetData)
	}
}

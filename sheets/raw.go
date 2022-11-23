package sheet

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"google.golang.org/api/sheets/v4"
)

func returnAllSheets(c *gin.Context) {
	id := c.Param("id")
	srv, _ := c.MustGet("service").(*sheets.Service)
	s := SheetService{srv: srv, id: id}
	data, err := s.getAllSheets()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

func returnSheetData(c *gin.Context) {
	id := c.Param("id")
	sheetRange := c.Param("range")
	srv, _ := c.MustGet("service").(*sheets.Service)
	s := SheetService{srv: srv, id: id}
	data, err := s.getSheetData(sheetRange)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

func updateSheetData(c *gin.Context) {
	id := c.Param("id")
	sheetRange := c.Param("range")
	srv, _ := c.MustGet("service").(*sheets.Service)

	var d RawData
	c.BindJSON(&d)

	s := SheetService{srv: srv, id: id}
	err := s.updateValue(sheetRange, d.Values)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Writer.WriteHeader(200)
}

package sheet

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/samber/lo"

	"google.golang.org/api/sheets/v4"
)

type GinContext struct {
	c *gin.Context
}

type Search struct {
	key    string
	values []string
}

func (gc GinContext) throwError(err error) {
	if err != nil {
		gc.c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		gc.c.Abort()
	}
}

func stringToInt64(sheetId string) int64 {
	result, err := strconv.ParseInt(sheetId, 10, 64)
	if err != nil {
		result = 0
	}
	return result
}

func formatData(keys []interface{}, data [][]interface{}) []map[string]string {
	result := lo.Map(data, func(item []interface{}, _ int) map[string]string {
		r := make(map[string]string)
		lo.ForEach(keys, func(key interface{}, index int) {
			if index < len(item) {
				r[key.(string)] = item[index].(string)
			} else {
				r[key.(string)] = ""
			}
		})
		return r
	})
	return result
}

func formatValue(keys []interface{}, data map[string]interface{}) [][]interface{} {
	v := make([]interface{}, len(keys))
	lo.ForEach(keys, func(key interface{}, index int) {
		if key != nil && data[key.(string)] != nil {
			v[index] = data[key.(string)]
		}
	})
	return [][]interface{}{v}
}

func updateItem(s SheetService, itemId int64, d map[string]interface{}) error {
	keys, err := s.getKeys(s.sheetId)
	data, err := s.getRowById(itemId)
	sheetRange := data.ValueRanges[0].ValueRange

	values := formatValue(keys, d)
	s.updateValue(sheetRange.Range, values)
	return err
}

func returnJsonSheetKeys(c *gin.Context) {
	id := c.Param("id")
	srv, _ := c.MustGet("service").(*sheets.Service)
	sheetId := stringToInt64(c.Param("sheetId"))

	s := SheetService{srv: srv, id: id, sheetId: sheetId}
	gc := GinContext{c: c}
	data, err := s.getKeys(sheetId)

	gc.throwError(err)
	c.JSON(http.StatusOK, data)
}

func returnSingleItem(c *gin.Context) {
	id := c.Param("id")
	srv, _ := c.MustGet("service").(*sheets.Service)
	sheetId := stringToInt64(c.Param("sheetId"))
	itemId := stringToInt64(c.Param("itemId"))

	s := SheetService{srv: srv, id: id, sheetId: sheetId}
	gc := GinContext{c: c}
	data, err := s.getRowById(itemId)
	keys, err := s.getKeys(sheetId)

	gc.throwError(err)
	c.JSON(http.StatusOK, formatData(keys, data.ValueRanges[0].ValueRange.Values))
}

func returnSearchData(c *gin.Context) {
	id := c.Param("id")
	srv, _ := c.MustGet("service").(*sheets.Service)
	sheetId := stringToInt64(c.Param("sheetId"))

	s := SheetService{srv: srv, id: id, sheetId: sheetId}
	gc := GinContext{c: c}
	data, err := s.getSheetDataById(sheetId)
	keys, err := s.getKeys(sheetId)
	query := c.Request.URL.Query()

	r := lo.Filter(formatData(keys, data), func(item map[string]string, index int) bool {
		return lo.EveryBy(lo.Keys(query), func(q string) bool {
			qv := lo.Flatten(lo.Map(query[q], func(value string, _ int) []string {
				return strings.Split(value, ",")
			}))
			return lo.Contains(qv, item[q])
		})
	})

	gc.throwError(err)
	c.JSON(http.StatusOK, r)
}

func returnJsonSheetData(c *gin.Context) {
	id := c.Param("id")
	srv, _ := c.MustGet("service").(*sheets.Service)
	sheetId := stringToInt64(c.Param("sheetId"))

	s := SheetService{srv: srv, id: id, sheetId: sheetId}
	gc := GinContext{c: c}
	data, err := s.getSheetDataById(sheetId)
	keys, err := s.getKeys(sheetId)

	gc.throwError(err)
	c.JSON(http.StatusOK, formatData(keys, data))
}

func updateSingleItem(c *gin.Context) {
	id := c.Param("id")
	srv, _ := c.MustGet("service").(*sheets.Service)
	sheetId := stringToInt64(c.Param("sheetId"))
	itemId := stringToInt64(c.Param("itemId"))
	s := SheetService{srv: srv, id: id, sheetId: sheetId}
	gc := GinContext{c: c}

	var d map[string]interface{}
	c.BindJSON(&d)

	err := updateItem(s, itemId, d)

	gc.throwError(err)
	c.Writer.WriteHeader(200)
}

func appendSingleItem(c *gin.Context) {
	id := c.Param("id")
	srv, _ := c.MustGet("service").(*sheets.Service)
	sheetId := stringToInt64(c.Param("sheetId"))

	s := SheetService{srv: srv, id: id, sheetId: sheetId}
	gc := GinContext{c: c}

	var d map[string]interface{}
	c.BindJSON(&d)

	if d["id"] == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Must include id"})
		return
	}

	itemId := stringToInt64(d["id"].(string))
	row, err := s.getRowById(itemId)
	gc.throwError(err)

	if row.ValueRanges == nil {
		rowNum, err := s.appendItem()
		gc.throwError(err)
		_, err = s.createMetaData(itemId, *rowNum)
		gc.throwError(err)
	}
	err = updateItem(s, itemId, d)

	gc.throwError(err)
	c.Writer.WriteHeader(200)
}

func deleteSingleItem(c *gin.Context) {
	id := c.Param("id")
	srv, _ := c.MustGet("service").(*sheets.Service)
	sheetId := stringToInt64(c.Param("sheetId"))
	itemId := stringToInt64(c.Param("itemId"))

	s := SheetService{srv: srv, id: id, sheetId: sheetId}
	gc := GinContext{c: c}

	_, err := s.deleteRow(itemId)

	gc.throwError(err)
	c.Writer.WriteHeader(200)
}

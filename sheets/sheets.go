package sheet

import (
	"fmt"
	"strconv"

	"github.com/samber/lo"

	"google.golang.org/api/sheets/v4"
)

type SheetService struct {
	srv     *sheets.Service
	id      string
	sheetId int64
}

func (s SheetService) getAllSheets() ([]*sheets.SheetProperties, error) {
	resp, err := s.srv.Spreadsheets.Get(s.id).Do()
	if err != nil {
		return nil, err
	}
	sheets := lo.Map(resp.Sheets, func(sheet *sheets.Sheet, index int) *sheets.SheetProperties {
		return sheet.Properties
	})
	return sheets, nil
}

func (s SheetService) getSheetData(sheetRange string) (*sheets.ValueRange, error) {
	resp, err := s.srv.Spreadsheets.Values.Get(s.id, sheetRange).Do()
	return resp, err
}

func (s SheetService) updateValue(sheetRange string, values [][]interface{}) error {
	vr := sheets.ValueRange{
		MajorDimension: "ROWS",
		Values:         values,
	}
	_, err := s.srv.Spreadsheets.Values.Update(s.id, sheetRange, &vr).ValueInputOption("RAW").Do()
	return err
}

func (s SheetService) getMetaData(id int64) (*sheets.DeveloperMetadata, error) {
	data, err := s.srv.Spreadsheets.DeveloperMetadata.Get(s.id, id).Do()
	return data, err
}

func (s SheetService) getValueByDataFilter(dataFilters []*sheets.DataFilter) (*sheets.BatchGetValuesByDataFilterResponse, error) {
	rq := &sheets.BatchGetValuesByDataFilterRequest{
		DataFilters: dataFilters,
	}
	data, err := s.srv.Spreadsheets.Values.BatchGetByDataFilter(s.id, rq).Do()
	return data, err
}

func (s SheetService) getSheetDataById(id int64) ([][]interface{}, error) {
	data, err := s.getValueByDataFilter([]*sheets.DataFilter{{
		GridRange: &sheets.GridRange{
			SheetId: id,
		},
	}})
	if err != nil {
		return nil, err
	}

	value := data.ValueRanges[0].ValueRange.Values
	return lo.Drop(value, 1), err
}

func (s SheetService) getKeys(id int64) ([]interface{}, error) {
	data, err := s.getValueByDataFilter([]*sheets.DataFilter{{
		GridRange: &sheets.GridRange{
			EndRowIndex: 1,
			SheetId:     id,
		},
	}})
	if err != nil {
		return nil, err
	}

	value := data.ValueRanges[0].ValueRange.Values
	return lo.Flatten(value), err
}

func (s SheetService) getRowById(id int64) (*sheets.BatchGetValuesByDataFilterResponse, error) {
	data, err := s.getValueByDataFilter([]*sheets.DataFilter{{
		DeveloperMetadataLookup: &sheets.DeveloperMetadataLookup{
			MetadataId: id,
		},
	}})
	return data, err
}

func (s SheetService) deleteRow(id int64) (*sheets.BatchUpdateSpreadsheetResponse, error) {
	m, err := s.getMetaData(id)
	if err != nil {
		return nil, err
	}
	rq := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{{
			DeleteDimension: &sheets.DeleteDimensionRequest{
				Range: &sheets.DimensionRange{
					SheetId:    s.sheetId,
					Dimension:  "ROWS",
					StartIndex: m.Location.DimensionRange.StartIndex,
					EndIndex:   m.Location.DimensionRange.EndIndex,
				},
			},
		}},
	}
	data, err := s.srv.Spreadsheets.BatchUpdate(s.id, rq).Do()
	return data, err
}

func (s SheetService) createMetaData(id int64, index int64) (*sheets.BatchUpdateSpreadsheetResponse, error) {
	rq := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{{
			CreateDeveloperMetadata: &sheets.CreateDeveloperMetadataRequest{
				DeveloperMetadata: &sheets.DeveloperMetadata{
					MetadataId:    id,
					MetadataKey:   "id",
					MetadataValue: strconv.FormatInt(id, 10),
					Visibility:    "DOCUMENT",
					Location: &sheets.DeveloperMetadataLocation{
						DimensionRange: &sheets.DimensionRange{
							Dimension:  "ROWS",
							StartIndex: index,
							EndIndex:   index + 1,
							SheetId:    s.sheetId,
						},
					},
				},
			},
		}},
	}
	data, err := s.srv.Spreadsheets.BatchUpdate(s.id, rq).Do()
	return data, err
}

func (s SheetService) appendItem() (*int64, error) {
	allSheets, err := s.getAllSheets()
	if err != nil {
		return nil, err
	}

	var sheetName string
	sheet, ok := lo.Find(allSheets, func(item *sheets.SheetProperties) bool {
		return item.Index == s.sheetId
	})

	if s.sheetId == 0 || !ok {
		sheetName = allSheets[0].Title
	} else {
		sheetName = sheet.Title
	}
	fmt.Print(sheetName)

	id := []interface{}{"id"}

	vr := sheets.ValueRange{
		MajorDimension: "ROWS",
		Values:         [][]interface{}{id},
	}

	result, err := s.srv.Spreadsheets.Values.Append(s.id, "a!A:A", &vr).ValueInputOption("RAW").Do()
	if err != nil {
		return nil, err
	}

	rr := result.Updates.UpdatedRange
	data, err := s.srv.Spreadsheets.Get(s.id).Ranges(rr).IncludeGridData(true).Do()
	if err != nil {
		return nil, err
	}

	rowNum := data.Sheets[0].Data[0].StartRow

	return &rowNum, err
}

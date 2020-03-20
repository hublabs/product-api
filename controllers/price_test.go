package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"github.com/hublabs/product-api/models"
	"testing"

	"github.com/labstack/echo"
	"github.com/pangpanglabs/goutils/test"
)

func TestPriceCRUD(t *testing.T) {
	inputs := []PriceInput{
		{
			Name:      "price#1",
			Barcode:   "barcode#1",
			SalePrice: 198.05,
		},
		{
			Name:      "price#2",
			ProductId: 1,
			SalePrice: 160,
		},
		{
			Name:      "price#3",
			Barcode:   "barcode#2",
			SalePrice: 179.02,
		},
	}

	for i, p := range inputs {
		pb, _ := json.Marshal(p)
		t.Run(fmt.Sprint("Create#", i+1), func(t *testing.T) {
			req := httptest.NewRequest(echo.POST, "/v1/prices", bytes.NewReader(pb))
			setHeader(req)
			rec := httptest.NewRecorder()
			test.Ok(t, handleWithFilter(PriceController{}.Create, echoApp.NewContext(req, rec)))
			test.Equals(t, http.StatusOK, rec.Code)
		})
	}

	t.Run("GetAllBarcode", func(t *testing.T) {
		req := httptest.NewRequest(echo.GET, "/v1/prices/barcode", nil)
		setHeader(req)
		rec := httptest.NewRecorder()
		test.Ok(t, handleWithFilter(PriceController{}.GetAllBarcode, echoApp.NewContext(req, rec)))
		test.Equals(t, http.StatusOK, rec.Code)

		var v struct {
			Result struct {
				TotalCount int                   `json:"totalCount"`
				Items      []models.PriceSkuInfo `json:"items"`
			} `json:"result"`
			Success bool `json:"success"`
		}
		test.Ok(t, json.Unmarshal(rec.Body.Bytes(), &v))
		test.Equals(t, v.Result.TotalCount, 2)
		test.Equals(t, v.Result.Items[0].TargetId, "barcode#2")
		test.Equals(t, v.Result.Items[0].SalePrice, 179.02)
		test.Equals(t, v.Result.Items[1].TargetId, "barcode#1")
		test.Equals(t, v.Result.Items[1].SalePrice, 198.05)
	})
}

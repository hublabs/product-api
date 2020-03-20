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

func TestBrandCRUD(t *testing.T) {
	inputs := []map[string]interface{}{
		{
			"code": "EE",
			"name": "Eland",
		},
		{
			"code": "EA",
			"name": "Eland Accessory",
		},
	}

	for i, p := range inputs {
		pb, _ := json.Marshal(p)
		t.Run(fmt.Sprint("Create#", i+1), func(t *testing.T) {
			req := httptest.NewRequest(echo.POST, "/v1/brands", bytes.NewReader(pb))
			setHeader(req)
			rec := httptest.NewRecorder()
			test.Ok(t, handleWithFilter(BrandController{}.Create, echoApp.NewContext(req, rec)))
			test.Equals(t, http.StatusOK, rec.Code)
		})
	}

	t.Run("GetAll", func(t *testing.T) {
		req := httptest.NewRequest(echo.GET, "/v1/brands", nil)
		setHeader(req)
		rec := httptest.NewRecorder()
		test.Ok(t, handleWithFilter(BrandController{}.GetAll, echoApp.NewContext(req, rec)))
		test.Equals(t, http.StatusOK, rec.Code)

		var v struct {
			Result struct {
				TotalCount int            `json:"totalCount"`
				Items      []models.Brand `json:"items"`
			} `json:"result"`
			Success bool `json:"success"`
		}
		test.Ok(t, json.Unmarshal(rec.Body.Bytes(), &v))
		test.Equals(t, v.Result.TotalCount, 2)
		test.Equals(t, v.Result.Items[0].Code, "EE")
	})

	t.Run("GetOne", func(t *testing.T) {
		req := httptest.NewRequest(echo.GET, "/v1/brands/1", nil)
		setHeader(req)
		rec := httptest.NewRecorder()
		c := echoApp.NewContext(req, rec)
		c.SetPath("/v1/brands/:id")
		c.SetParamNames("id")
		c.SetParamValues("1")
		test.Ok(t, handleWithFilter(BrandController{}.GetOne, c))
		test.Equals(t, http.StatusOK, rec.Code)

		var v struct {
			Result  models.Brand `json:"result"`
			Success bool         `json:"success"`
		}
		test.Ok(t, json.Unmarshal(rec.Body.Bytes(), &v))
		test.Equals(t, v.Result.Code, "EE")
	})
}

package controllers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hublabs/product-api/models"

	"github.com/labstack/echo"
	"github.com/pangpanglabs/goutils/test"
)

func TestSkuCRUD(t *testing.T) {
	t.Run("GetAll", func(t *testing.T) {
		req := httptest.NewRequest(echo.GET, "/v1/skus?brandIds=1,2", nil)
		setHeader(req)
		rec := httptest.NewRecorder()
		test.Ok(t, handleWithFilter(SkuController{}.GetAll, echoApp.NewContext(req, rec)))
		test.Equals(t, http.StatusOK, rec.Code)

		var v struct {
			Result struct {
				TotalCount int          `json:"totalCount"`
				Items      []models.Sku `json:"items"`
			} `json:"result"`
			Success bool `json:"success"`
		}
		test.Ok(t, json.Unmarshal(rec.Body.Bytes(), &v))
		test.Equals(t, v.Result.TotalCount, 2)
		test.Equals(t, v.Result.Items[0].Name, "sku#2")
	})

	t.Run("GetAllWithItem", func(t *testing.T) {
		req := httptest.NewRequest(echo.GET, "/v1/skus?brandIds=1,2&fields=item&withOffer=true", nil)
		setHeader(req)
		rec := httptest.NewRecorder()
		test.Ok(t, handleWithFilter(SkuController{}.GetAll, echoApp.NewContext(req, rec)))
		test.Equals(t, http.StatusOK, rec.Code)

		var v struct {
			Result struct {
				TotalCount int          `json:"totalCount"`
				Items      []models.Sku `json:"items"`
			} `json:"result"`
			Success bool `json:"success"`
		}
		test.Ok(t, json.Unmarshal(rec.Body.Bytes(), &v))
		test.Equals(t, v.Result.TotalCount, 2)
		test.Equals(t, v.Result.Items[0].Name, "sku#2")
		test.Equals(t, v.Result.Items[1].Name, "sku#1")
	})

	t.Run("GetOne", func(t *testing.T) {
		req := httptest.NewRequest(echo.GET, "/v1/skus/1", nil)
		setHeader(req)
		rec := httptest.NewRecorder()
		c := echoApp.NewContext(req, rec)
		c.SetPath("/v1/skus/:id")
		c.SetParamNames("id")
		c.SetParamValues("1")
		test.Ok(t, handleWithFilter(SkuController{}.GetOne, c))
		test.Equals(t, http.StatusOK, rec.Code)

		var v struct {
			Result  models.Sku `json:"result"`
			Success bool       `json:"success"`
		}
		test.Ok(t, json.Unmarshal(rec.Body.Bytes(), &v))
		test.Equals(t, v.Result.Name, "sku#1")
	})
}

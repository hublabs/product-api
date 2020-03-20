package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hublabs/product-api/models"

	"github.com/labstack/echo"
	"github.com/pangpanglabs/goutils/test"
)

func TestProductCRUD(t *testing.T) {
	inputs := []map[string]interface{}{
		{
			"name": "product#1",
			"brand": map[string]interface{}{
				"id": 1,
			},
			"skus": []map[string]interface{}{
				{
					"name": "sku#1",
				},
			},
			"attributes": map[string]string{
				"Year": "2018",
			},
			"listPrice": 200,
		},
		{
			"name": "product#2",
			"brand": map[string]interface{}{
				"id": 2,
			},
			"skus": []map[string]interface{}{
				{
					"name": "sku#2",
				},
			},
			"attributes": map[string]string{
				"Year": "2019",
			},
			"listPrice": 100,
		},
	}

	for i, p := range inputs {
		pb, _ := json.Marshal(p)
		t.Run(fmt.Sprint("Create#", i+1), func(t *testing.T) {
			req := httptest.NewRequest(echo.POST, "/v1/products", bytes.NewReader(pb))
			setHeader(req)
			rec := httptest.NewRecorder()
			test.Ok(t, handleWithFilter(ProductController{}.CreateOrUpdate, echoApp.NewContext(req, rec)))
			test.Equals(t, http.StatusOK, rec.Code)
		})
	}

	t.Run("GetAll", func(t *testing.T) {
		req := httptest.NewRequest(echo.GET, "/v1/products?brandIds=1,2", nil)
		setHeader(req)
		rec := httptest.NewRecorder()
		test.Ok(t, handleWithFilter(ProductController{}.GetAll, echoApp.NewContext(req, rec)))
		test.Equals(t, http.StatusOK, rec.Code)

		var v struct {
			Result struct {
				TotalCount int              `json:"totalCount"`
				Items      []models.Product `json:"items"`
			} `json:"result"`
			Success bool `json:"success"`
		}
		test.Ok(t, json.Unmarshal(rec.Body.Bytes(), &v))
		test.Equals(t, v.Result.TotalCount, 2)
		test.Equals(t, v.Result.Items[0].Name, "product#2")
		test.Equals(t, len(v.Result.Items[0].Attributes), 0)
	})

	t.Run("GetAllWithAttribute", func(t *testing.T) {
		req := httptest.NewRequest(echo.GET, "/v1/products?brandIds=1,2&fields=attribute", nil)
		setHeader(req)
		rec := httptest.NewRecorder()
		test.Ok(t, handleWithFilter(ProductController{}.GetAll, echoApp.NewContext(req, rec)))
		test.Equals(t, http.StatusOK, rec.Code)

		var v struct {
			Result struct {
				TotalCount int              `json:"totalCount"`
				Items      []models.Product `json:"items"`
			} `json:"result"`
			Success bool `json:"success"`
		}
		test.Ok(t, json.Unmarshal(rec.Body.Bytes(), &v))
		test.Equals(t, v.Result.TotalCount, 2)
		test.Equals(t, v.Result.Items[0].Name, "product#2")
		test.Equals(t, len(v.Result.Items[0].Attributes), 1)
	})

	t.Run("GetAllWithItem", func(t *testing.T) {
		req := httptest.NewRequest(echo.GET, "/v1/products?brandIds=1,2&fields=sku&fields=item&withOffer=true", nil)
		setHeader(req)
		rec := httptest.NewRecorder()
		test.Ok(t, handleWithFilter(ProductController{}.GetAll, echoApp.NewContext(req, rec)))
		test.Equals(t, http.StatusOK, rec.Code)

		var v struct {
			Result struct {
				TotalCount int              `json:"totalCount"`
				Items      []models.Product `json:"items"`
			} `json:"result"`
			Success bool `json:"success"`
		}
		test.Ok(t, json.Unmarshal(rec.Body.Bytes(), &v))
		test.Equals(t, v.Result.TotalCount, 2)
		test.Equals(t, v.Result.Items[0].Name, "product#2")
		test.Equals(t, len(v.Result.Items[0].Skus), 1)

		test.Equals(t, v.Result.Items[1].Name, "product#1")
		test.Equals(t, len(v.Result.Items[1].Skus), 1)
	})

	var product models.Product

	t.Run("GetOne", func(t *testing.T) {
		req := httptest.NewRequest(echo.GET, "/v1/products/1", nil)
		setHeader(req)
		rec := httptest.NewRecorder()
		c := echoApp.NewContext(req, rec)
		c.SetPath("/v1/products/:id")
		c.SetParamNames("id")
		c.SetParamValues("1")
		test.Ok(t, handleWithFilter(ProductController{}.GetOne, c))
		test.Equals(t, http.StatusOK, rec.Code)

		var v struct {
			Result  models.Product `json:"result"`
			Success bool           `json:"success"`
		}
		test.Ok(t, json.Unmarshal(rec.Body.Bytes(), &v))
		test.Equals(t, v.Result.Name, "product#1")
		product = v.Result
	})

	t.Run("Update", func(t *testing.T) {
		product.Name = "product#updated"
		pb, _ := json.Marshal(product)
		req := httptest.NewRequest(echo.POST, "/v1/products", bytes.NewReader(pb))
		setHeader(req)
		rec := httptest.NewRecorder()
		c := echoApp.NewContext(req, rec)
		c.SetPath("/v1/products/:id")
		c.SetParamNames("id")
		c.SetParamValues(fmt.Sprintf("%v", product.Id))
		test.Ok(t, handleWithFilter(ProductController{}.CreateOrUpdate, c))
		test.Equals(t, http.StatusOK, rec.Code)

		var v struct {
			Result  models.Product `json:"result"`
			Success bool           `json:"success"`
		}
		test.Ok(t, json.Unmarshal(rec.Body.Bytes(), &v))
		test.Equals(t, v.Result.Name, "product#updated")
	})
}

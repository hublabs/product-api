package controllers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/hublabs/common/api"
	"github.com/hublabs/product-api/models"

	"github.com/labstack/echo"
	"github.com/pangpanglabs/echoswagger"
	"github.com/pangpanglabs/goutils/converter"
)

type SkuController struct{}

func (c SkuController) Init(g echoswagger.ApiGroup) {
	g.SetSecurity("Authorization")

	g.GET("", c.GetAll).
		AddParamQueryNested(GetAllSkuInput{})
	g.GET("/:id", c.GetOne).
		AddParamPath(0, "id", "Id of Sku").
		AddParamQueryNested(FieldAndStoreInput{})
	// According to https://stackoverflow.com/questions/5020704/how-to-design-restful-search-filtering
	// `/searches` with POST method should be a standard of search/filter resources with long parameter.

	// [Updated 20190724] After discuss with jang.jaehue, add a GET method for same API
	// refer to https://gitlab.srxcloud.com/github.com/hublabs/product-api/merge_requests/64
	// https://www.elastic.co/guide/en/elasticsearch/reference/current/search-request-body.html
	g.GET("/uids", c.GetByUids).
		AddParamBody(SearchSkuByUidInput{}, "body", "", true)
	g.POST("/uids", c.GetByUids).
		AddParamBody(SearchSkuByUidInput{}, "body", "", true)
	g.GET("/searches", c.SearchAll).
		AddParamBody(SearchSkuInput{}, "body", "", true)
	g.POST("/searches", c.SearchAll).
		AddParamBody(SearchSkuInput{}, "body", "", true)
}

func (SkuController) GetAll(c echo.Context) error {
	var v GetAllSkuInput
	if err := c.Bind(&v); err != nil {
		return api.ErrorParameter.New(err)
	}
	if err := c.Validate(&v); err != nil {
		return api.ErrorParameter.New(err)
	}
	ids := converter.StringToIntSlice(v.Ids)
	brandIds := converter.StringToIntSlice(v.BrandIds)
	codes := converter.StringToStringSlice(v.Codes)
	if code := strings.TrimSpace(v.Code); code != "" {
		codes = append(codes, code)
	}
	if v.Q == "" && len(ids) == 0 && len(brandIds) == 0 && v.BrandCode == "" && len(codes) == 0 && v.ProductCode == "" && v.Barcode == "" {
		return api.ErrorMissParameter.New(errors.New("at least one parameter: q, ids, brandIds, brandCode, code, codes, productCode, barcode"))
	}

	if v.MaxResultCount == 0 {
		v.MaxResultCount = defaultMaxResultCount
	}
	hasMore, totalCount, skus, err := models.Sku{}.GetAll(c.Request().Context(), v.Q, v.ProductCode, v.Barcode, v.BrandCode, v.Enable, v.Saleable, codes, ids, brandIds, v.SkipCount, v.MaxResultCount, v.Sortby, v.Order, v.Fields, v.WithHasMore)
	if err != nil {
		return api.ErrorDB.New(err)
	}
	return renderSuccArray(c, v.WithHasMore, hasMore, totalCount, skus)
}

func (SkuController) GetOne(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return api.ErrorParameter.New(err)
	}
	var v FieldAndStoreInput
	if err := c.Bind(&v); err != nil {
		return api.ErrorParameter.New(err)
	}

	sku, err := models.Sku{}.GetOne(c.Request().Context(), id, v.Fields)
	if err != nil {
		return api.ErrorDB.New(err)
	}
	if sku == nil {
		return api.ErrorNotFound.New(nil)
	}
	return renderSucc(c, http.StatusOK, sku)
}

func (SkuController) SearchAll(c echo.Context) error {
	var v SearchSkuInput
	if err := c.Bind(&v); err != nil {
		return api.ErrorParameter.New(err)
	}
	if err := c.Validate(&v); err != nil {
		return api.ErrorParameter.New(err)
	}
	if v.Q == "" && len(v.Filters) == 0 {
		return api.ErrorMissParameter.New(errors.New("at least one parameter: q, filters"))
	}
	if v.MaxResultCount == 0 {
		v.MaxResultCount = defaultMaxResultCount
	}
	hasMore, totalCount, skus, err := models.Sku{}.SearchAll(c.Request().Context(), v.Q, v.Enable, v.Saleable, v.Filters, v.SkipCount, v.MaxResultCount, v.Sortby, v.Order, v.Fields, v.WithHasMore)
	if err != nil {
		return api.ErrorDB.New(err)
	}
	return renderSuccArray(c, v.WithHasMore, hasMore, totalCount, skus)
}

func (SkuController) GetByUids(c echo.Context) error {
	var v SearchSkuByUidInput
	if err := c.Bind(&v); err != nil {
		return api.ErrorParameter.New(err)
	}

	result, err := models.Sku{}.GetByUids(c.Request().Context(), v.Source, v.Fields, v.Uids...)
	if err != nil {
		return api.ErrorDB.New(err)
	}

	return renderSucc(c, http.StatusOK, result)
}

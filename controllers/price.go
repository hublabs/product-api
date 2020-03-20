package controllers

import (
	"net/http"

	"github.com/hublabs/common/api"
	"github.com/hublabs/product-api/models"

	"github.com/labstack/echo"
	"github.com/pangpanglabs/echoswagger"
)

type PriceController struct{}

func (c PriceController) Init(g echoswagger.ApiGroup) {
	g.SetSecurity("Authorization")

	// Portal商品页创建销售价需要
	g.POST("", c.Create).
		AddParamBody(PriceInput{}, "body", "PriceInput model", true)
	// 查询未登记商品
	g.GET("/barcode", c.GetAllBarcode).
		AddParamQueryNested(SearchInput{})
}

func (PriceController) Create(c echo.Context) error {
	var v PriceInput
	if err := c.Bind(&v); err != nil {
		return renderFail(c, api.ErrorParameter.New(err))
	}

	p := v.ToModel()
	if err := p.Create(c.Request().Context()); err != nil {
		return renderFail(c, api.ErrorDB.New(err))
	}

	return renderSucc(c, http.StatusOK, p)
}

func (PriceController) GetAllBarcode(c echo.Context) error {
	var v SearchInput
	if err := c.Bind(&v); err != nil {
		return renderFail(c, api.ErrorParameter.New(err))
	}
	if v.MaxResultCount == 0 {
		v.MaxResultCount = defaultMaxResultCount
	}
	totalCount, prices, err := models.Price{}.GetAllBarcode(c.Request().Context(), v.SkipCount, v.MaxResultCount, v.Sortby, v.Order)
	if err != nil {
		return renderFail(c, api.ErrorDB.New(err))
	}
	return renderSuccArray(c, false, false, totalCount, prices)
}

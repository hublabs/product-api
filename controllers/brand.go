package controllers

import (
	"net/http"
	"strconv"

	"github.com/hublabs/common/api"
	"github.com/hublabs/product-api/models"

	"github.com/labstack/echo"
	"github.com/pangpanglabs/echoswagger"
	"github.com/pangpanglabs/goutils/converter"
)

type BrandController struct{}

func (c BrandController) Init(g echoswagger.ApiGroup) {
	g.SetSecurity("Authorization")

	g.GET("", c.GetAll).
		AddParamQueryNested(GetAllBrandInput{})
	g.GET("/:id", c.GetOne).
		AddParamPath(0, "id", "Id of Brand")
	g.POST("", c.Create).
		AddParamBody(models.Brand{}, "body", "Brand model", true)
	g.PUT("/:id", c.Update).
		AddParamPath(0, "id", "Id of Brand").
		AddParamBody(models.Brand{}, "body", "Brand model", true)
}

func (BrandController) GetAll(c echo.Context) error {
	var v GetAllBrandInput
	if err := c.Bind(&v); err != nil {
		return renderFail(c, api.ErrorParameter.New(err))
	}
	if v.MaxResultCount == 0 {
		v.MaxResultCount = defaultMaxResultCount
	}

	ids := converter.StringToIntSlice(v.Ids)
	codes := converter.StringToStringSlice(v.Codes)

	totalCount, brands, err := models.Brand{}.GetAll(c.Request().Context(), v.Q, v.Code, v.Enable, ids, codes, v.SkipCount, v.MaxResultCount)
	if err != nil {
		return renderFail(c, api.ErrorDB.New(err))
	}
	return renderSuccArray(c, false, false, totalCount, brands)
}

func (BrandController) GetOne(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return renderFail(c, api.ErrorParameter.New(err))
	}

	brand, err := models.Brand{}.GetById(c.Request().Context(), id)
	if err != nil {
		return renderFail(c, api.ErrorDB.New(err))
	}

	if brand == nil {
		return renderFail(c, api.ErrorNotFound.New(err))
	}
	return renderSucc(c, http.StatusOK, brand)
}

func (BrandController) Create(c echo.Context) error {
	var brand models.Brand
	if err := c.Bind(&brand); err != nil {
		return renderFail(c, api.ErrorParameter.New(err))
	}
	if err := brand.Create(c.Request().Context()); err != nil {
		return renderFail(c, api.ErrorDB.New(err))
	}
	return renderSucc(c, http.StatusOK, brand)
}

func (BrandController) Update(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return renderFail(c, api.ErrorParameter.New(err))
	}

	var brand models.Brand
	if err := c.Bind(&brand); err != nil {
		return renderFail(c, api.ErrorParameter.New(err))
	}

	brand.Id = id
	if err := brand.Update(c.Request().Context()); err != nil {
		return renderFail(c, api.ErrorDB.New(err))
	}
	return renderSucc(c, http.StatusOK, brand)
}

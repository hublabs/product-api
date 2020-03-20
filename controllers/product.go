package controllers

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/hublabs/common/api"
	"github.com/hublabs/product-api/models"

	"github.com/360EntSecGroup-Skylar/excelize/v2"
	"github.com/labstack/echo"
	"github.com/pangpanglabs/echoswagger"
	"github.com/pangpanglabs/goutils/converter"
)

type ProductController struct{}

func (c ProductController) Init(g echoswagger.ApiGroup) {
	g.SetSecurity("Authorization")

	g.GET("", c.GetAll).
		AddParamQueryNested(GetAllProductInput{})
	g.GET("/:id", c.GetOne).
		AddParamPath(0, "id", "Id of Product").
		AddParamQueryNested(FieldAndStoreInput{})
	g.POST("", c.CreateOrUpdate).
		AddParamBody(models.Product{}, "body", "Product model", true)
	g.GET("/searches", c.SearchAll).
		AddParamBody(SearchProductInput{}, "body", "", true)
	g.POST("/searches", c.SearchAll).
		AddParamBody(SearchProductInput{}, "body", "", true)
	g.POST("/validate-excel", c.ValidateImportExcel).
		AddParamFile("file", "excel", true)
	g.POST("/batch", c.BatchImport).
		AddParamBody([]models.ProductImportTemplate{}, "body", "ProductImportTemplate model", true)
	g.GET("/statistics", c.StatisticsData)
}

func (ProductController) GetAll(c echo.Context) error {
	var v GetAllProductInput
	if err := c.Bind(&v); err != nil {
		return renderFail(c, api.ErrorParameter.New(err))
	}
	if err := c.Validate(&v); err != nil {
		return renderFail(c, api.ErrorParameter.New(err))
	}
	ids := converter.StringToIntSlice(v.Ids)
	brandIds := converter.StringToIntSlice(v.BrandIds)
	codes := converter.StringToStringSlice(v.Codes)
	if code := strings.TrimSpace(v.Code); code != "" {
		codes = append(codes, code)
	}
	if v.Q == "" && len(ids) == 0 && len(brandIds) == 0 && v.BrandCode == "" && len(codes) == 0 {
		return renderFail(c, api.ErrorMissParameter.New(errors.New("at least one parameter: q, ids, brandIds, brandCode, code, codes")))
	}
	if v.MaxResultCount == 0 {
		v.MaxResultCount = defaultMaxResultCount
	}
	hasMore, totalCount, products, err := models.Product{}.GetAll(c.Request().Context(), v.Q, v.HasDigital, v.HasTitleImage, v.BrandCode, v.Enable, codes, ids, brandIds, v.SkipCount, v.MaxResultCount, v.Sortby, v.Order, v.Fields, v.WithHasMore)
	if err != nil {
		return renderFail(c, api.ErrorDB.New(err))
	}
	return renderSuccArray(c, v.WithHasMore, hasMore, totalCount, products)
}

func (ProductController) GetOne(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return renderFail(c, api.ErrorParameter.New(err))
	}
	var v FieldAndStoreInput
	if err := c.Bind(&v); err != nil {
		return renderFail(c, api.ErrorParameter.New(err))
	}
	product, err := models.Product{}.GetOne(c.Request().Context(), id, v.Fields)
	if err != nil {
		return renderFail(c, api.ErrorDB.New(err))
	}
	if product == nil {
		return renderFail(c, api.ErrorNotFound.New(nil))
	}
	return renderSucc(c, http.StatusOK, product)
}

func (ProductController) CreateOrUpdate(c echo.Context) error {
	var product models.Product
	if err := c.Bind(&product); err != nil {
		return renderFail(c, api.ErrorParameter.New(err))
	}
	result, err := models.Product{}.CreateOrUpdate(c.Request().Context(), product)
	if err != nil {
		return renderFail(c, api.ErrorDB.New(err))
	}
	return renderSucc(c, http.StatusOK, result)
}

func (ProductController) SearchAll(c echo.Context) error {
	var v SearchProductInput
	if err := c.Bind(&v); err != nil {
		return renderFail(c, api.ErrorParameter.New(err))
	}
	if err := c.Validate(&v); err != nil {
		return renderFail(c, api.ErrorParameter.New(err))
	}
	if v.Q == "" && len(v.Filters) == 0 {
		return renderFail(c, api.ErrorMissParameter.New(errors.New("at least one parameter: q, filters")))
	}

	if v.MaxResultCount == 0 {
		v.MaxResultCount = defaultMaxResultCount
	}
	hasMore, totalCount, products, err := models.Product{}.SearchAll(c.Request().Context(), v.Q, v.Enable, v.Filters, v.SkipCount, v.MaxResultCount, v.Sortby, v.Order, v.Fields, v.WithHasMore)
	if err != nil {
		return renderFail(c, api.ErrorDB.New(err))
	}

	return renderSuccArray(c, v.WithHasMore, hasMore, totalCount, products)
}

func (ProductController) ValidateImportExcel(c echo.Context) error {
	file, err := c.FormFile("file")
	if err != nil {
		return renderFail(c, api.ErrorParameter.New(err))
	}
	data, err := file.Open()
	if err != nil {
		return renderFail(c, api.ErrorParameter.New(err))
	}
	defer data.Close()

	xlsx, err := excelize.OpenReader(data)
	if err != nil {
		return renderFail(c, api.ErrorParameter.New(err))
	}

	rows, err := xlsx.GetRows("商品")
	if err != nil {
		return renderFail(c, api.ErrorParameter.New(err))
	}
	var pList []models.ProductImportTemplate
	for i := range rows {
		if i == 0 {
			continue
		}
		var p models.ProductImportTemplate
		var listPrice, salePrice float64
		if rows[i][0] == "" {
			p.ErrorList = append(p.ErrorList, 10001) //商品编码
		}
		if rows[i][1] == "" {
			p.ErrorList = append(p.ErrorList, 10002) //sku编号
		}
		if rows[i][2] == "" {
			p.ErrorList = append(p.ErrorList, 10014) //品牌名称
		}
		if rows[i][3] == "" {
			p.ErrorList = append(p.ErrorList, 10015) //品牌Code
		}
		if rows[i][4] == "" {
			p.ErrorList = append(p.ErrorList, 10003) //商品名称
		}
		listPrice, _ = strconv.ParseFloat(rows[i][7], 64)
		if listPrice <= 0 {
			p.ErrorList = append(p.ErrorList, 10006) //吊牌价
		}
		salePrice, _ = strconv.ParseFloat(rows[i][8], 64)
		if salePrice <= 0 {
			p.ErrorList = append(p.ErrorList, 10007) //销售价
		}
		if rows[i][7] == "" {
			p.ErrorList = append(p.ErrorList, 10014)
		}
		if rows[i][8] == "" {
			p.ErrorList = append(p.ErrorList, 10015)
		}
		p.ProductCode = rows[i][0]
		p.SkuCode = rows[i][1]
		p.BrandName = rows[i][2]
		p.BrandCode = rows[i][3]
		p.ProductName = rows[i][4]
		p.Color = rows[i][5]
		p.Size = rows[i][6]
		p.SkuName = p.Color + "|" + p.Size
		p.ListPrice = listPrice
		p.SalePrice = salePrice
		if listPrice < salePrice {
			p.ErrorList = append(p.ErrorList, 10008)
		}
		if p.SkuCode == "" && p.ProductCode == "" && p.ProductName == "" && p.Color == "" && p.Size == "" && p.BrandCode == "" && p.BrandName == "" {
			continue
		}
		pList = append(pList, p)
	}
	for i := range pList {
		var lcount, scount, rcount int
		for j := range pList {
			if pList[i].BrandCode == pList[j].BrandCode && pList[i].ProductCode == pList[j].ProductCode && pList[i].ListPrice != pList[j].ListPrice {
				lcount++
			}
			if pList[i].BrandCode == pList[j].BrandCode && pList[i].ProductCode == pList[j].ProductCode && pList[i].SalePrice != pList[j].SalePrice {
				scount++
			}
			if pList[i].BrandCode == pList[j].BrandCode && pList[i].SkuCode == pList[j].SkuCode && i != j {
				rcount++
			}
		}
		if lcount > 0 {
			pList[i].ErrorList = append(pList[i].ErrorList, 10011)
		}
		if scount > 0 {
			pList[i].ErrorList = append(pList[i].ErrorList, 10012)
		}
		if rcount > 0 {
			pList[i].ErrorList = append(pList[i].ErrorList, 10013)
		}
	}

	list, err := models.ProductImportTemplate{}.ValidateImport(c.Request().Context(), pList)
	if err != nil {
		return renderFail(c, api.ErrorDB.New(err))
	}
	return renderSucc(c, http.StatusOK, list)
}

func (ProductController) BatchImport(c echo.Context) error {
	var list []models.ProductImportTemplate
	if err := c.Bind(&list); err != nil {
		return renderFail(c, api.ErrorParameter.New(err))
	}
	ctx := context.WithValue(c.Request().Context(), models.DataSourceContext, models.DataSourceExcel)
	result, err := models.ProductImportTemplate{}.BatchImport(ctx, list)
	if err != nil {
		return renderFail(c, api.ErrorDB.New(err))
	}
	return renderSucc(c, http.StatusOK, result)
}

func (ProductController) StatisticsData(c echo.Context) error {
	result, err := (models.Product{}).StatisticsData(c.Request().Context())
	if err != nil {
		return renderFail(c, api.ErrorDB.New(err))
	}
	return renderSucc(c, http.StatusOK, result)
}

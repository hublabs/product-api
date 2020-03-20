package controllers

import (
	"strconv"

	"github.com/hublabs/product-api/models"
)

const (
	defaultMaxResultCount = 30
)

type PagingInput struct {
	SkipCount      int `json:"skipCount" query:"skipCount"`
	MaxResultCount int `json:"maxResultCount" query:"maxResultCount" valid:"range(0|500)"`
}

type SortingInput struct {
	Sortby []string `json:"sortby" query:"sortby"`
	Order  []string `json:"order" query:"order"`
}

type SearchInput struct {
	PagingInput
	SortingInput
}

type FieldAndStoreInput struct {
	Fields    models.FieldTypeList `json:"fields" query:"fields"`
	StoreId   int64                `json:"storeId" query:"storeId"`
	WithOffer bool                 `json:"withOffer" query:"withOffer"`
}

type GetAllBrandInput struct {
	Q      string `query:"q"`
	Code   string `query:"code"`
	Codes  string `query:"codes"`
	Ids    string `query:"ids"`
	Enable string `query:"enable"`
	PagingInput
}

type GetAllProductInput struct {
	Q             string `query:"q" valid:"stringlength(3|64)"`
	Code          string `query:"code"`
	Codes         string `query:"codes"`
	BrandCode     string `query:"brandCode"`
	HasDigital    string `query:"hasDigital"`
	HasTitleImage string `query:"hasTitleImage"`
	Ids           string `query:"ids"`
	BrandIds      string `query:"brandIds"`
	Enable        string `query:"enable"`
	WithHasMore   bool   `query:"withHasMore"`
	FieldAndStoreInput
	SearchInput
}

type GetAllSkuInput struct {
	Q             string `query:"q" valid:"stringlength(3|64)"`
	Code          string `query:"code"`
	Codes         string `query:"codes"`
	BrandCode     string `query:"brandCode"`
	ProductCode   string `query:"productCode"`
	Barcode       string `query:"barcode"`
	HasDigital    string `query:"hasDigital"`
	HasTitleImage string `query:"hasTitleImage"`
	ActionIndex   int    `query:"actionIndex"`
	TargetNo      int    `query:"targetNo"`
	Ids           string `query:"ids"`
	BrandIds      string `query:"brandIds"`
	Enable        string `query:"enable"`
	Saleable      string `query:"saleable"`
	WithHasMore   bool   `query:"withHasMore"`
	FieldAndStoreInput
	SearchInput
}

type SearchProductInput struct {
	Q           string               `json:"q" valid:"stringlength(3|64)"`
	Enable      string               `json:"enable"`
	Filters     models.Filter        `json:"filters"`
	Fields      models.FieldTypeList `json:"fields"`
	WithHasMore bool                 `json:"withHasMore"`
	SearchInput
}

type SearchSkuInput struct {
	Q           string        `json:"q" valid:"stringlength(3|64)"`
	Enable      string        `json:"enable"`
	Filters     models.Filter `json:"filters"`
	Saleable    string        `json:"saleable"`
	WithHasMore bool          `json:"withHasMore"`
	FieldAndStoreInput
	SearchInput
}

type SearchSkuByUidInput struct {
	Uids   []string             `json:"uids"`
	Source string               `json:"source"`
	Fields models.FieldTypeList `json:"fields"`
}

type PriceInput struct {
	ProductId int64   `json:"productId"`
	Barcode   string  `json:"barcode"`
	SalePrice float64 `json:"salePrice"`
	Name      string  `json:"name"`
}

type ItemInput struct {
	Barcode   string  `json:"barcode"`
	SalePrice float64 `json:"salePrice"`
}

func (p PriceInput) ToModel() models.Price {
	v := models.Price{
		SalePrice: p.SalePrice,
	}
	switch {
	case p.ProductId != 0:
		v.TargetType = models.PriceTargetTypeProduct
		v.TargetId = strconv.FormatInt(p.ProductId, 10)
	case p.Barcode != "":
		v.TargetType = models.PriceTargetTypeBarcode
		v.TargetId = p.Barcode
	}
	return v
}

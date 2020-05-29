package models

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/hublabs/common/auth"

	"github.com/go-xorm/xorm"
)

const IdentifierSourceBarcode = "Barcode"

type FieldType string

const (
	FieldTypeProduct   FieldType = "product"
	FieldTypeSku       FieldType = "sku"
	FieldTypeAttribute FieldType = "attribute"
)

type FieldTypeList []FieldType

func (l FieldTypeList) Contains(f ...FieldType) bool {
	for _, t := range f {
		for _, lf := range l {
			if t == lf {
				return true
			}
		}
	}
	return false
}

func (l *FieldTypeList) Add(f ...FieldType) {
	*l = append(*l, f...)
}

type ComparerType string

const (
	ComparerTypeInclude          ComparerType = "in"
	ComparerTypeNotInclude       ComparerType = "nin"
	ComparerTypeGreaterThanEqual ComparerType = "gte"
	ComparerTypeLessThanEqual    ComparerType = "lte"
	ComparerTypeBetween          ComparerType = "between"
)

const (
	ConditionTypeBrandCode           = "brand_code"
	ConditionTypeProduct             = "product_id"
	ConditionTypeProductCode         = "product_code"
	ConditionTypeSkuCode             = "sku_code"
	ConditionTypeBarcode             = "barcode"
	ConditionTypeListPrice           = "list_price"
	ConditionTypeAttributeYear       = "year"
	ConditionTypeAttributeSeasonCode = "season_code"
	ConditionTypeAttributeItemCode   = "item_code"
	ConditionTypeAttributeYearSeason = "year_season_code"
	ConditionTypeAttributeSaleMonth  = "sale_month"
)

type Filter map[string]FilterItem

type FilterItem struct {
	Comparer ComparerType `json:"comparer"`
	Values   []string     `json:"values"`
}

func IsValidConditionType(t string) bool {
	switch t {
	case ConditionTypeBrandCode,
		ConditionTypeProduct,
		ConditionTypeProductCode,
		ConditionTypeSkuCode,
		ConditionTypeBarcode,
		ConditionTypeListPrice,
		ConditionTypeAttributeYear,
		ConditionTypeAttributeSeasonCode,
		ConditionTypeAttributeYearSeason,
		ConditionTypeAttributeSaleMonth:
		return true
	}
	return false
}

func setSortOrder(q xorm.Interface, sortby, order []string, table ...string) error {
	connect := func(col string) string {
		if len(table) > 0 {
			return table[0] + "." + col
		}
		return col
	}

	if len(sortby) != 0 {
		if len(sortby) == len(order) {
			// 1) for each sort field, there is an associated order
			for i, v := range sortby {
				v = connect(v)
				if order[i] == "desc" {
					q.Desc(v)
				} else if order[i] == "asc" {
					q.Asc(v)
				} else if order[i] != "" {
					return errors.New("Invalid order. Must be either [asc|desc]")
				}
			}
		} else if len(sortby) != len(order) && len(order) == 1 {
			// 2) there is exactly one order, all the sorted fields will be sorted by this order
			for _, v := range sortby {
				v = connect(v)
				if order[0] == "desc" {
					q.Desc(v)
				} else if order[0] == "asc" {
					q.Asc(v)
				} else if order[0] != "" {
					return errors.New("Invalid order. Must be either [asc|desc]")
				}
			}
		} else if len(sortby) != len(order) && len(order) != 1 {
			return errors.New("'sortby', 'order' sizes mismatch or 'order' size is not 1")
		}
	} else {
		if len(order) != 0 && order[0] != "" {
			return errors.New("unused 'order' fields")
		}
	}
	return nil
}

func excludeDeleted(table string) string {
	return fmt.Sprintf("`%s`.`deleted_at` IS NULL OR `%s`.`deleted_at`= '0001-01-01 00:00:00'", table, table)
}

func tenantCode(ctx context.Context) string {
	user := auth.UserClaim{}.FromCtx(ctx)
	return user.TenantCode
}

func appendStrArgs(args []interface{}, isNum bool, ts ...string) []interface{} {
	for _, t := range ts {
		if isNum {
			f, _ := strconv.ParseFloat(t, 64)
			args = append(args, f)
		} else {
			args = append(args, t)
		}
	}
	return args
}

func placeholder(length int) string {
	placeholder := strings.Repeat("?,", length)
	return placeholder[:len(placeholder)-1]
}

func (p ProductImportTemplate) ToProduct(tenantCode string, brandId int64) (Product, error) {
	var identifiers []SkuIdentifier
	if p.BarCode != "" {
		identifiers = append(identifiers, SkuIdentifier{
			Uid:    p.BarCode,
			Source: "Barcode",
		})
	}
	var options []Option
	options = append(options, Option{
		Name:  "color",
		Value: p.Color,
	})
	options = append(options, Option{
		Name:  "size",
		Value: p.Size,
	})
	sku := Sku{
		Name:        p.SkuName,
		Code:        p.SkuCode,
		Identifiers: identifiers,
		Options:     options,
	}
	price := Price{
		TenantCode: tenantCode,
		TargetType: PriceTargetTypeProduct,
		SalePrice:  p.SalePrice,
	}

	return Product{
		Name:       p.ProductName,
		Code:       p.ProductCode,
		TenantCode: tenantCode,
		BrandId:    brandId,
		ListPrice:  p.ListPrice,
		Prices:     []Price{price},
		Skus:       []Sku{sku},
		HasDigital: false,
		Enable:     true,
	}, nil
}

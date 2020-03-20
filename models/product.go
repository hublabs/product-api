package models

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/hublabs/product-api/adapters"
	"github.com/hublabs/product-api/factory"

	"github.com/go-xorm/xorm"
)

type Product struct {
	Id           int64               `json:"id"`
	TenantCode   string              `json:"-" xorm:"index varchar(16)"`
	Code         string              `json:"code" xorm:"index varchar(64)"`
	Name         string              `json:"name"`
	BrandId      int64               `json:"-" xorm:"index"`
	Brand        Brand               `json:"brand" xorm:"-"`
	TitleImage   string              `json:"titleImage"`
	ListPrice    float64             `json:"listPrice"`
	SupGroupCode string              `json:"-"`
	Prices       []Price             `json:"prices,omitempty" xorm:"-"`
	Identifiers  []ProductIdentifier `json:"identifiers,omitempty" xorm:"-"`
	Skus         []Sku               `json:"skus,omitempty" xorm:"-"`
	Attributes   map[string]string   `json:"attributes,omitempty" xorm:"-"`
	HasDigital   bool                `json:"hasDigital" xorm:"index"`
	Enable       bool                `json:"enable" xorm:"index"`
	CreatedAt    time.Time           `json:"createdAt" xorm:"created"`
	UpdatedAt    time.Time           `json:"updatedAt" xorm:"updated"`
	DeletedAt    time.Time           `json:"-" xorm:"deleted index"`
}

type ProductImportTemplate struct {
	ProductName string  `json:"productName"`
	ProductCode string  `json:"productCode"`
	SkuCode     string  `json:"skuCode"`
	SkuName     string  `json:"skuName"`
	Color       string  `json:"color"`
	Size        string  `json:"size"`
	ListPrice   float64 `json:"listPrice"`
	SalePrice   float64 `json:"salePrice"`
	BarCode     string  `json:"barCode"`
	ErrorList   []int   `json:"errorList"`
	Status      string  `json:"status"`
	BrandCode   string  `json:"brandCode"`
	BrandName   string  `json:"brandName"`
}

// Must be private because of event ProductCreated
func (p *Product) create(ctx context.Context) (err error) {
	_, err = factory.DB(ctx).Insert(p)
	return
}

// Must be private because of event ProductChanged
func (p *Product) update(ctx context.Context, hasDigital bool) (err error) {
	cols := []string{
		"code", "name", "list_price", "brand_id",
	}
	if p.TitleImage != "" {
		cols = append(cols, "title_image")
	}
	if hasDigital {
		cols = append(cols, "has_digital")
	}
	if _, err = factory.DB(ctx).ID(p.Id).Cols(cols...).Update(p); err != nil {
		return err
	}

	for i := range p.Identifiers {
		p.Identifiers[i].ProductId = p.Id
		if err := p.Identifiers[i].CreateOrUpdate(ctx); err != nil {
			return err
		}
	}

	if err := (AttributeValue{}).CreateOrUpdates(ctx, p.Id, p.Attributes); err != nil {
		return err
	}

	if err := p.removeIdentifiersExcept(ctx, p.Identifiers); err != nil {
		return err
	}
	prices, err := Price{}.GetByTarget(ctx, PriceTargetTypeProduct, strconv.FormatInt(p.Id, 10))

PriceLoop:
	for k, _ := range p.Prices {
		for j, _ := range prices {
			if p.Prices[k].SalePrice == prices[j].SalePrice {
				// TODO 潜在的问题，如果price: 100 -> 200 -> 100，最后一次价格无法生成
				continue PriceLoop
			}
		}
		p.Prices[k].TargetId = strconv.FormatInt(p.Id, 10)
		if err := p.Prices[k].Create(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (Product) GetOne(ctx context.Context, id int64, fields FieldTypeList) (*Product, error) {
	var products ProductList
	if err := factory.DB(ctx).Where("id = ?", id).Limit(1).Find(&products); err != nil {
		return nil, err
	} else if len(products) == 0 {
		return nil, nil
	}

	if err := products.LoadPrices(ctx); err != nil {
		return nil, err
	}

	if err := products.LoadIdentifiers(ctx); err != nil {
		return nil, err
	}

	if fields.Contains(FieldTypeAttribute) {
		if err := products.LoadAttributes(ctx); err != nil {
			return nil, err
		}
	}

	if err := products.LoadSkus(ctx); err != nil {
		return nil, err
	}

	if err := products.LoadBrands(ctx); err != nil {
		return nil, err
	}

	return &products[0], nil
}

func (Product) GetAll(ctx context.Context, q, hasDigital, hasTitleImage, brandCode, enable string, codes []string, ids, brandIds []int64, skipCount, maxResultCount int, sortby, order []string, fields FieldTypeList, withHasMore bool) (bool, int64, []Product, error) {
	query := factory.DB(ctx).Where("tenant_code = ?", tenantCode(ctx))
	if len(sortby) == 0 || len(order) == 0 {
		sortby = []string{"id"}
		order = []string{"desc"}
	}
	if err := setSortOrder(query, sortby, order); err != nil {
		return false, 0, nil, err
	}

	if len(ids) != 0 {
		query.In("product.id", ids)
	}

	if len(brandIds) != 0 {
		query.In("product.brand_id", brandIds)
	}

	if brandCode != "" {
		query.Where("product.brand_id IN (SELECT id FROM brand WHERE brand.code = ?)", brandCode)
	}

	if hasDigital == "has" {
		query.Where("product.has_digital = ?", true)
	} else if hasDigital == "not" {
		query.Where("product.has_digital = ?", false)
	}

	if hasTitleImage == "has" {
		query.Where("product.title_image <> ?", "")
	} else if hasTitleImage == "not" {
		query.Where("product.title_image = ?", "")
	}

	if enable != "" {
		b, _ := strconv.ParseBool(enable)
		query.Where("product.enable = ?", b)
	}

	if len(codes) != 0 {
		query.In("product.code", codes)
	}

	if q != "" {
		query.Where("product.code LIKE ?", q+"%")
	}

	var (
		products   ProductList
		hasMore    bool
		totalCount int64
		err        error
	)

	if withHasMore {
		err = query.Limit(maxResultCount+1, skipCount).Find(&products)
		if len(products) == maxResultCount+1 {
			products = products[:maxResultCount]
			hasMore = true
		}
	} else {
		totalCount, err = query.Limit(maxResultCount, skipCount).FindAndCount(&products)
	}
	if err != nil {
		return false, 0, nil, err
	}

	if len(products) == 0 {
		return false, 0, nil, nil
	}

	if err := products.LoadPrices(ctx); err != nil {
		return false, 0, nil, err
	}

	if err := products.LoadIdentifiers(ctx); err != nil {
		return false, 0, nil, err
	}

	if fields.Contains(FieldTypeAttribute) {
		if err := products.LoadAttributes(ctx); err != nil {
			return false, 0, nil, err
		}
	}

	if fields.Contains(FieldTypeSku) {
		if err := products.LoadSkus(ctx); err != nil {
			return false, 0, nil, err
		}
	}

	if err := products.LoadBrands(ctx); err != nil {
		return false, 0, nil, err
	}

	return hasMore, totalCount, products, nil
}

func (Product) CreateOrUpdate(ctx context.Context, product Product) (*Product, error) {
	var p Product
	product.TenantCode = tenantCode(ctx)
	exist, err := factory.DB(ctx).Where("id = ?", product.Id).Get(&p)
	if err != nil {
		return nil, err
	}
	product.BrandId = product.Brand.Id
	if !exist || p.Id == 0 {
		if err := product.create(ctx); err != nil {
			return nil, err
		}
		for i := range product.Identifiers {
			product.Identifiers[i].ProductId = product.Id
			if err := product.Identifiers[i].create(ctx); err != nil {
				return nil, err
			}
		}
		for k, v := range product.Attributes {
			if _, err := (AttributeValue{}).Create(ctx, product.Id, k, v); err != nil {
				return nil, err
			}
		}
		for k := range product.Prices {
			product.Prices[k].TargetId = strconv.FormatInt(product.Id, 10)
			if err := product.Prices[k].Create(ctx); err != nil {
				return nil, err
			}
		}
		for j := range product.Skus {
			product.Skus[j].ProductId = product.Id
			if err := product.Skus[j].Create(ctx); err != nil {
				return nil, err
			}
		}
		if err := (adapters.MessagePublisher{}).Publish(ctx, product, adapters.EventProductCreated); err != nil {
			return nil, err
		}
		return &product, nil
	}

	if err := product.update(ctx, true); err != nil {
		return nil, err
	}
	var skus []Sku
	if err := factory.DB(ctx).Where("product_id = ?", product.Id).Find(&skus); err != nil {
		return nil, err
	}
SkuLoop:
	for i, sku := range product.Skus {
		sku.ProductId = p.Id
		for j := range skus {
			if sku.Id == skus[j].Id {
				if err := sku.Update(ctx); err != nil {
					return nil, err
				}
				continue SkuLoop
			}
		}
		if err := sku.Create(ctx); err != nil {
			return nil, nil
		}
		product.Skus[i].Id = sku.Id
	}
	if err := p.removeSkusExcept(ctx, product.Skus); err != nil {
		return nil, err
	}
	if err := (adapters.MessagePublisher{}).Publish(ctx, product, adapters.EventProductChanged); err != nil {
		return nil, err
	}
	return &product, nil
}

// 不删除以前Product下的Sku而现在不存在的数据
func (Product) GetByCode(ctx context.Context, code string) (*Product, error) {
	var p Product
	_, err := factory.DB(ctx).Where("code = ?", code).Get(&p)
	return &p, err
}

func (Product) GetByIdentifier(ctx context.Context, identifier, source string, brandId int64) (*Product, bool, error) {
	var p Product
	exist, err := factory.DB(ctx).Where("product.id in (SELECT product_id from product_identifier WHERE uid = ? AND source = ?)", identifier, source).And("brand_id = ?", brandId).Get(&p)
	if err != nil {
		return nil, false, err
	}
	if exist {
		return &p, true, nil
	}
	return nil, false, nil
}

func (p Product) removeIdentifiersExcept(ctx context.Context, except []ProductIdentifier) (err error) {
	var exceptIdentifilerIds []int64
	for _, identifier := range except {
		exceptIdentifilerIds = append(exceptIdentifilerIds, identifier.Id)
	}
	_, err = factory.DB(ctx).Where("product_id = ?", p.Id).NotIn("id", exceptIdentifilerIds).Delete(&ProductIdentifier{})
	return
}

func (p Product) removeSkusExcept(ctx context.Context, except []Sku) error {
	var exceptSkuIds []int64
	for _, s := range except {
		exceptSkuIds = append(exceptSkuIds, s.Id)
	}

	var removeSkuIds []int64
	if err := factory.DB(ctx).Table("sku").Select("id").
		Where("product_id = ?", p.Id).NotIn("id", exceptSkuIds).
		Find(&removeSkuIds); err != nil {
		return err
	}

	if len(removeSkuIds) == 0 {
		return nil
	}

	if _, err := factory.DB(ctx).In("id", removeSkuIds).Delete(&Sku{}); err != nil {
		return err
	}
	if _, err := factory.DB(ctx).In("sku_id", removeSkuIds).Delete(&Option{}); err != nil {
		return err
	}
	if _, err := factory.DB(ctx).In("sku_id", removeSkuIds).Delete(&SkuIdentifier{}); err != nil {
		return err
	}

	return nil
}

type ProductList []Product

func (products ProductList) Ids() (ids []interface{}) {
	for _, s := range products {
		ids = append(ids, s.Id)
	}
	return
}

func (products ProductList) Find(id int64) *Product {
	for i := range products {
		if products[i].Id == id {
			return &products[i]
		}
	}
	return nil
}

func (products ProductList) LoadSkus(ctx context.Context) error {
	var skus SkuList
	if err := factory.DB(ctx).In("product_id", products.Ids()...).Find(&skus); err != nil {
		return err
	}

	if len(skus) == 0 {
		return nil
	}

	if err := skus.LoadIdentifiers(ctx); err != nil {
		return err
	}

	if err := skus.LoadOptions(ctx); err != nil {
		return err
	}

	// set Product
	for _, s := range skus {
		p := products.Find(s.ProductId)
		if p != nil {
			p.Skus = append(p.Skus, s)
		}
	}
	return nil
}

func (products ProductList) LoadPrices(ctx context.Context) error {
	var prices []Price
	if err := factory.DB(ctx).
		Where("target_type = ?", PriceTargetTypeProduct).
		In("target_id", products.Ids()...).
		Desc("id").
		Find(&prices); err != nil {
		return err
	}
	for _, price := range prices {
		i, err := strconv.ParseInt(price.TargetId, 10, 64)
		if err != nil {
			return err
		}
		productId := i
		if productId != 0 {
			p := products.Find(productId)
			if p != nil {
				p.Prices = append(p.Prices, price)
			}
		}
	}

	for i := range products {
		if len(products[i].Prices) == 0 {
			products[i].Prices = []Price{
				{SalePrice: products[i].ListPrice},
			}
		} else {
			products[i].Prices = []Price{products[i].Prices[0]}
		}
	}

	return nil
}

func (products ProductList) LoadIdentifiers(ctx context.Context) error {
	var identifiers []ProductIdentifier
	if err := factory.DB(ctx).In("product_id", products.Ids()...).Find(&identifiers); err != nil {
		return err
	}
	for _, identifier := range identifiers {
		s := products.Find(identifier.ProductId)
		if s != nil {
			s.Identifiers = append(s.Identifiers, identifier)
		}
	}
	return nil
}

func (products ProductList) LoadBrands(ctx context.Context) error {
	var brandIds []interface{}
	for _, p := range products {
		brandIds = append(brandIds, p.BrandId)
	}

	var brands []Brand
	if err := factory.DB(ctx).In("id", brandIds...).Find(&brands); err != nil {
		return err
	}

	findBrand := func(brandId int64) *Brand {
		for _, b := range brands {
			if b.Id == brandId {
				return &b
			}
		}
		return nil
	}

	for i := range products {
		p := &products[i]
		if b := findBrand(p.BrandId); b != nil {
			p.Brand = *b
		}
	}

	return nil
}

func (product *Product) LoadAttributes(ctx context.Context) error {
	var attrExtends []AttributeExtends
	if err := factory.DB(ctx).Table("attribute_value").Select("attribute.*, attribute_value.*").
		Join("INNER", "attribute", "attribute_value.attribute_id = attribute.id").
		Where("attribute_value.product_id = ?", product.Id).Find(&attrExtends); err != nil {
		return err
	}
	product.Attributes = make(map[string]string)
	for _, attr := range attrExtends {
		product.Attributes[attr.Name] = attr.Value
	}
	return nil
}

func (products ProductList) LoadAttributes(ctx context.Context) error {
	var attrExtends []AttributeExtends
	if err := factory.DB(ctx).Table("attribute_value").Select("attribute.*, attribute_value.*").
		Join("INNER", "attribute", "attribute_value.attribute_id = attribute.id").
		In("attribute_value.product_id", products.Ids()...).Find(&attrExtends); err != nil {
		return err
	}
	for _, attr := range attrExtends {
		s := products.Find(attr.ProductId)
		if s == nil {
			continue
		}
		if s.Attributes == nil {
			s.Attributes = make(map[string]string)
		}
		s.Attributes[attr.Name] = attr.Value
	}
	return nil
}

func (Product) SearchAll(ctx context.Context, q, enable string, filter Filter, skipCount, maxResultCount int, sortby, order []string, fields FieldTypeList, withHasMore bool) (bool, int64, []Product, error) {
	query := factory.DB(ctx).Where("tenant_code = ?", tenantCode(ctx))
	if q != "" {
		query.Where("code LIKE ?", q+"%")
	}
	if enable != "" {
		b, _ := strconv.ParseBool(enable)
		query.Where("enable = ?", b)
	}
	filterQuery(query, filter)

	var (
		products   ProductList
		hasMore    bool
		totalCount int64
		err        error
	)

	if len(sortby) == 0 || len(order) == 0 {
		sortby = []string{"id"}
		order = []string{"desc"}
	}

	if err := setSortOrder(query, sortby, order); err != nil {
		return false, 0, nil, err
	}

	if withHasMore {
		err = query.Limit(maxResultCount+1, skipCount).Find(&products)
		if len(products) == maxResultCount+1 {
			products = products[:maxResultCount]
			hasMore = true
		}
	} else {
		totalCount, err = query.Limit(maxResultCount, skipCount).FindAndCount(&products)
	}
	if err != nil {
		return false, 0, nil, err
	}

	if len(products) == 0 {
		return false, 0, nil, nil
	}

	if err := products.LoadPrices(ctx); err != nil {
		return false, 0, nil, err
	}

	if err := products.LoadIdentifiers(ctx); err != nil {
		return false, 0, nil, err
	}

	if fields.Contains(FieldTypeAttribute) {
		if err := products.LoadAttributes(ctx); err != nil {
			return false, 0, nil, err
		}
	}

	if fields.Contains(FieldTypeSku) {
		if err := products.LoadSkus(ctx); err != nil {
			return false, 0, nil, err
		}
	}

	if err := products.LoadBrands(ctx); err != nil {
		return false, 0, nil, err
	}

	return hasMore, totalCount, products, nil
}

func filterQuery(query *xorm.Session, filter Filter) {
	condQuery := func(c ComparerType, v []string, conditionType string) (string, string, []interface{}) {
		var isNum bool
		if conditionType == ConditionTypeListPrice || conditionType == ConditionTypeProduct {
			isNum = true
		}
		var args []interface{}
		switch c {
		case ComparerTypeNotInclude:
			return "NOT EXISTS", fmt.Sprintf("IN (%s)", placeholder(len(v))), appendStrArgs(args, isNum, v...)
		case ComparerTypeGreaterThanEqual:
			return "IN", ">= ?", appendStrArgs(args, isNum, v[0])
		case ComparerTypeLessThanEqual:
			return "IN", "<= ?", appendStrArgs(args, isNum, v[0])
		case ComparerTypeBetween:
			return "IN", "BETWEEN ? AND ?", appendStrArgs(args, isNum, v[0], v[1])
		default:
			return "IN", fmt.Sprintf("IN (%s)", placeholder(len(v))), appendStrArgs(args, isNum, v...)
		}
	}

	for k, v := range filter {
		if len(v.Values) == 0 {
			continue
		}
		keyword, clause, args := condQuery(v.Comparer, v.Values, k)
		switch k {
		case ConditionTypeBrandCode:
			if v.Comparer == ComparerTypeNotInclude {
				query.And(fmt.Sprintf(`%v (SELECT 1 FROM brand WHERE product.brand_id = brand.id AND brand.code %v)`, keyword, clause), args...)
			} else {
				query.And(fmt.Sprintf(`product.brand_id %v (SELECT id FROM brand WHERE brand.code %v)`, keyword, clause), args...)
			}
		case ConditionTypeListPrice:
			if v.Comparer == ComparerTypeNotInclude {
				query.And(fmt.Sprintf(`%v (SELECT 1 FROM product AS p WHERE product.id = p.id AND p.list_price %v)`, keyword, clause), args...)
			} else {
				query.And(fmt.Sprintf(`product.list_price %v`, clause), args...)
			}
		case ConditionTypeProduct:
			if v.Comparer == ComparerTypeNotInclude {
				query.And(fmt.Sprintf(`%v (SELECT 1 FROM product AS p WHERE product.id = p.id AND p.id %v)`, keyword, clause), args...)
			} else {
				query.And(fmt.Sprintf(`product.id %v`, clause), args...)
			}
		case ConditionTypeProductCode:
			if v.Comparer == ComparerTypeNotInclude {
				query.And(fmt.Sprintf(`%v (SELECT 1 FROM product AS p WHERE product.id = p.id AND p.code %v)`, keyword, clause), args...)
			} else {
				query.And(fmt.Sprintf(`product.code %v`, clause), args...)
			}
		case ConditionTypeSkuCode:
			if v.Comparer == ComparerTypeNotInclude {
				query.And(fmt.Sprintf(`%v (SELECT 1 FROM sku WHERE product.id = sku.product_id AND sku.code %v)`, keyword, clause), args...)
			} else {
				query.And(fmt.Sprintf(`product.id %v (SELECT product_id FROM sku WHERE code %v)`, keyword, clause), args...)
			}
		case ConditionTypeBarcode:
			if v.Comparer == ComparerTypeNotInclude {
				query.And(fmt.Sprintf(`%v (SELECT 1 FROM sku AS s JOIN sku_identifier AS si ON s.id = si.sku_id WHERE product.id = s.product_id AND si.source = 'Barcode' AND si.uid %v)`, keyword, clause), args...)
			} else {
				query.And(fmt.Sprintf(`product.id %v (SELECT s.product_id FROM sku AS s JOIN sku_identifier AS si ON s.id = si.sku_id WHERE si.source = ? AND si.uid %v)`, keyword, clause), args...)
			}
		case ConditionTypeAttributeItemCode:
			if v.Comparer == ComparerTypeNotInclude {
				query.And(fmt.Sprintf(`%v (SELECT 1 FROM attribute AS a JOIN attribute_value AS av ON a.id = av.attribute_id WHERE product.id = av.product_id AND a.name = 'ItemCode' AND av.value %v)`, keyword, clause), args...)
			} else {
				query.And(fmt.Sprintf(`product.id %v (SELECT av.product_id FROM attribute AS a JOIN attribute_value AS av ON a.id = av.attribute_id WHERE a.name = 'ItemCode' AND av.value %v)`, keyword, clause), args...)
			}
		case ConditionTypeAttributeYear:
			if v.Comparer == ComparerTypeNotInclude {
				query.And(fmt.Sprintf(`%v (SELECT 1 FROM attribute AS a JOIN attribute_value AS av ON a.id = av.attribute_id WHERE product.id = av.product_id AND a.name = 'Year' AND av.value %v)`, keyword, clause), args...)
			} else {
				query.And(fmt.Sprintf(`product.id %v (SELECT av.product_id FROM attribute AS a JOIN attribute_value AS av ON a.id = av.attribute_id WHERE a.name = 'Year' AND av.value %v)`, keyword, clause), args...)
			}
		case ConditionTypeAttributeSeasonCode:
			if v.Comparer == ComparerTypeNotInclude {
				query.And(fmt.Sprintf(`%v (SELECT 1 FROM attribute AS a JOIN attribute_value AS av ON a.id = av.attribute_id WHERE product.id = av.product_id AND a.name = 'SeasonCode' AND av.value %v)`, keyword, clause), args...)
			} else {
				query.And(fmt.Sprintf(`product.id %v (SELECT av.product_id FROM attribute AS a JOIN attribute_value AS av ON a.id = av.attribute_id WHERE a.name = 'SeasonCode' AND av.value %v)`, keyword, clause), args...)
			}
		case ConditionTypeAttributeSaleMonth:
			if v.Comparer == ComparerTypeNotInclude {
				query.And(fmt.Sprintf(`%v (SELECT 1 FROM attribute AS a JOIN attribute_value AS av ON a.id = av.attribute_id WHERE product.id = av.product_id AND a.name = 'SaleMonth' AND av.value %v)`, keyword, clause), args...)
			} else {
				query.And(fmt.Sprintf(`product.id %v (SELECT av.product_id FROM attribute AS a JOIN attribute_value AS av ON a.id = av.attribute_id WHERE a.name = 'SaleMonth' AND av.value %v)`, keyword, clause), args...)
			}
		case ConditionTypeAttributeYearSeason:
			if v.Comparer == ComparerTypeNotInclude {
				query.And(fmt.Sprintf(`%v (SELECT 1 FROM attribute AS a JOIN attribute_value AS av ON a.id = av.attribute_id WHERE product.id = av.product_id AND a.name = 'YearSeasonCode' AND av.value %v)`, keyword, clause), args...)
			} else {
				query.And(fmt.Sprintf(`product.id %v (SELECT av.product_id FROM attribute AS a JOIN attribute_value AS av ON a.id = av.attribute_id WHERE a.name = 'YearSeasonCode' AND av.value %v)`, keyword, clause), args...)
			}
		}
	}
}

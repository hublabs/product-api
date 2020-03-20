package models

import (
	"context"
	"strconv"

	"github.com/hublabs/product-api/adapters"
	"github.com/hublabs/product-api/factory"
)

func (Product) CreateOrUpdateByCode(ctx context.Context, product Product) (*Product, error) {
	p, err := Product{}.GetByCode(ctx, product.Code)
	if err != nil {
		return nil, err
	}
	product.TenantCode = tenantCode(ctx)
	product.BrandId = product.Brand.Id
	if p == nil || p.Id == 0 {
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
	product.Id = p.Id
	if product.BrandId == 0 && p.BrandId != 0 {
		product.BrandId = p.BrandId
		brand, err := Brand{}.GetById(ctx, p.BrandId)
		if err != nil {
			return nil, err
		}
		product.Brand = *brand
	}

	if err := product.update(ctx, false); err != nil {
		return nil, err
	}
	skus, err := Sku{}.GetByProductId(ctx, p.Id)
	if err != nil {
		return nil, err
	}
SkuLoop:
	for i, sku := range product.Skus {
		sku.ProductId = p.Id
		for j := range skus {
			if sku.Code == skus[j].Code {
				sku.Id = skus[j].Id
				product.Skus[i].Id = skus[j].Id
				//如果skuCode存在，只更新sku的名字和option,identifiers
				if _, err = factory.DB(ctx).ID(sku.Id).Cols("name").Update(&sku); err != nil {
					return nil, err
				}
				for _, identifier := range product.Skus[i].Identifiers {
					identifier.SkuId = product.Skus[i].Id
					var d SkuIdentifier
					exist, err := factory.DB(ctx).Where("uid = ?", identifier.Uid).Get(&d)
					if err != nil {
						return nil, err
					}
					if !exist {
						if err := identifier.create(ctx); err != nil {
							return nil, err
						}
					} else {
						identifier.Id = d.Id
						if err := identifier.update(ctx); err != nil {
							return nil, err
						}
					}
				}
				for k := range product.Skus[i].Options {
					product.Skus[i].Options[k].SkuId = product.Skus[i].Id
					if err = product.Skus[i].Options[k].UpdateBySkuId(ctx); err != nil {
						return nil, err
					}
				}
				continue SkuLoop
			}
		}
		if err := sku.Create(ctx); err != nil {
			return nil, nil
		}
		product.Skus[i].Id = sku.Id
	}

	if err := (adapters.MessagePublisher{}).Publish(ctx, product, adapters.EventProductChanged); err != nil {
		return nil, err
	}
	return &product, nil
}

func (ProductImportTemplate) BatchImport(ctx context.Context, list []ProductImportTemplate) ([]Product, error) {
	var products []Product
	tenantCode := tenantCode(ctx)
	for i := range list {
		brand := &Brand{
			Code: list[i].BrandCode,
			Name: list[i].BrandName,
		}
		err := brand.GetOrCreate(ctx)
		if err != nil {
			return nil, err
		}
		p, err := list[i].ToProduct(tenantCode, brand.Id)
		if err != nil {
			return nil, err
		}
		p.Brand = *brand
		product, err := Product{}.CreateOrUpdateByCode(ctx, p)
		if err != nil {
			return nil, err
		}
		products = append(products, *product)
	}
	return products, nil
}

func (ProductImportTemplate) ValidateImport(ctx context.Context, list []ProductImportTemplate) ([]ProductImportTemplate, error) {
ProductLoop:
	for i := range list {
		var productList []struct {
			Product Product `xorm:"extends"`
			Sku     Sku     `xorm:"extends"`
		}
		if list[i].BrandCode != "" {
			if err := factory.DB(ctx).Table("product").Select("product.*,sku.*").
				Join("left", "brand", "brand.id = product.brand_id").
				Join("left", "sku", "sku.product_id = product.id").
				Where("sku.code = ?", list[i].SkuCode).
				And("brand.code = ?", list[i].BrandCode).Find(&productList); err != nil {
				return list, err
			}
			for j := range productList {
				if productList[j].Product.Code != list[i].ProductCode {
					list[i].ErrorList = append(list[i].ErrorList, 10009)
					continue ProductLoop
				}
			}
		}

		if len(list[i].ErrorList) == 0 {
			if len(productList) == 0 {
				list[i].Status = "Insert"
			} else {
				list[i].Status = "Update"
			}
		}
	}
	return list, nil
}

func (Product) StatisticsData(ctx context.Context) (interface{}, error) {
	type Data struct {
		BrandCode string `json:"brandCode" xorm:"brandCode"`
		BrandName string `json:"brandName" xorm:"brandName"`
		Count     int    `json:"count" xorm:"cnt"`
	}
	var list []Data
	if err := factory.DB(ctx).Table("product").Select("brand.code AS brandCode, brand.name AS brandName, COUNT(*) AS cnt").
		Join("INNER", "brand", "brand.id = product.brand_id").
		Where(excludeDeleted("product")).
		GroupBy("brand.code, brand.name").
		Find(&list); err != nil {
		return nil, err
	}

	count, err := factory.DB(ctx).Count(Product{})
	if err != nil {
		return nil, err
	}

	result := struct {
		Data       []Data `json:"data"`
		TotalCount int64  `json:"totalCount"`
	}{
		Data:       list,
		TotalCount: count,
	}

	return result, nil
}

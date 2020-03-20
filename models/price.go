package models

import (
	"context"
	"time"

	"github.com/hublabs/product-api/adapters"
	"github.com/hublabs/product-api/factory"
)

const (
	PriceTargetTypeProduct = "product"
	PriceTargetTypeBarcode = "barcode"
)

type PriceTargetType string

type Price struct {
	Id         int64           `json:"id"`
	TenantCode string          `json:"-" xorm:"index varchar(16)"`
	TargetType PriceTargetType `json:"targetType" xorm:"index"`
	TargetId   string          `json:"targetId" xorm:"index"`
	SalePrice  float64         `json:"salePrice"`
	CreatedAt  time.Time       `json:"createdAt" xorm:"created"`
	UpdatedAt  time.Time       `json:"updatedAt" xorm:"updated"`
}

type PriceSkuInfo struct {
	SkuId     int64      `json:"skuId" xorm:"-"`
	TargetId  string     `json:"targetId"`
	SalePrice float64    `json:"salePrice"`
	CreatedAt time.Time  `json:"createdAt"`
	MappedAt  *time.Time `json:"mappedAt,omitempty" xorm:"-"`
}

func (p *Price) Create(ctx context.Context) error {
	p.TenantCode = tenantCode(ctx)
	if _, err := factory.DB(ctx).Insert(p); err != nil {
		return err
	}
	return adapters.MessagePublisher{}.Publish(ctx, p, adapters.EventProductPriceChanged)
}

func (Price) Get(ctx context.Context, priceId int64) (*Price, error) {
	var p Price
	exist, err := factory.DB(ctx).ID(priceId).Get(&p)
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, nil
	}
	return &p, nil
}

func (Price) GetByTarget(ctx context.Context, targetType PriceTargetType, targetId string) ([]Price, error) {
	var prices []Price
	if err := factory.DB(ctx).
		Where("target_type = ?", targetType).And("target_id = ?", targetId).Desc("id").
		Find(&prices); err != nil {
		return nil, err
	}
	return prices, nil
}

func (Price) GetAllBarcode(ctx context.Context, skipCount, maxResultCount int, sortby, order []string) (int64, []PriceSkuInfo, error) {
	query := factory.DB(ctx).Table("price").Where("tenant_code = ? AND target_type = ?", tenantCode(ctx), PriceTargetTypeBarcode)
	if err := setSortOrder(query, sortby, order); err != nil {
		return 0, nil, err
	}
	var prices []PriceSkuInfo
	totalCount, err := query.Desc("price.id").Limit(maxResultCount, skipCount).FindAndCount(&prices)
	if err != nil {
		return 0, nil, err
	}

	if len(prices) == 0 {
		return 0, nil, nil
	}

	var uids []string
	for _, price := range prices {
		uids = append(uids, price.TargetId)
	}

	var identifiers []SkuIdentifier
	if err := factory.DB(ctx).In("uid", uids).Find(&identifiers); err != nil {
		return 0, nil, err
	}

	findSkuIdentifier := func(uid string) SkuIdentifier {
		for _, identifier := range identifiers {
			if uid == identifier.Uid {
				return identifier
			}
		}
		return SkuIdentifier{}
	}

	for i := range prices {
		identifier := findSkuIdentifier(prices[i].TargetId)
		if identifier.SkuId != 0 {
			prices[i].SkuId = identifier.SkuId
			prices[i].MappedAt = &identifier.CreatedAt
		}
	}

	return totalCount, prices, nil
}

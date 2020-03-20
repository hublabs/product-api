package models

import (
	"context"
	"github.com/hublabs/product-api/factory"
)

type Option struct {
	Id    int64  `json:"id"`
	SkuId int64  `json:"skuId" xorm:"index"`
	Code  string `json:"-"`
	Name  string `json:"name"`
	Value string `json:"value"`
}

// Must be private because of event SkuAdded、SkuChanged
func (o *Option) create(ctx context.Context) (err error) {
	_, err = factory.DB(ctx).Insert(o)
	return err
}

// Must be private because of event SkuChanged
func (o *Option) update(ctx context.Context) (err error) {
	cols := []string{
		"name", "code", "value",
	}

	if _, err = factory.DB(ctx).ID(o.Id).Cols(cols...).Update(o); err != nil {
		return err
	}
	return nil
}

func (o Option) UpdateBySkuId(ctx context.Context) error {
	if _, err := factory.DB(ctx).Cols("value").Where("name = ?", o.Name).And("sku_id = ?", o.SkuId).Update(&o); err != nil {
		return err
	}
	return nil
}

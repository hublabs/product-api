package models

import (
	"context"
	"time"

	"github.com/hublabs/product-api/factory"
)

type Attribute struct {
	Id   int64  `json:"id"`
	Name string `json:"name" xorm:"unique"`
}

type AttributeValue struct {
	Id          int64     `json:"id"`
	AttributeId int64     `json:"attributeId" xorm:"index"`
	ProductId   int64     `json:"productId" xorm:"index"`
	Value       string    `json:"value" xorm:"index"`
	CreatedAt   time.Time `json:"createdAt" xorm:"created"`
	UpdatedAt   time.Time `json:"updatedAt" xorm:"updated"`
}

type AttributeExtends struct {
	Attribute      `xorm:"extends"`
	AttributeValue `xorm:"extends"`
}

func (a *Attribute) create(ctx context.Context) error {
	_, err := factory.DB(ctx).Insert(a)
	return err
}

func (AttributeValue) Create(ctx context.Context, productId int64, key, value string) (*AttributeValue, error) {
	attr, err := Attribute{}.GetByName(ctx, key)
	if err != nil {
		return nil, err
	} else if attr == nil {
		// Create Attribute
		attr = &Attribute{
			Name: key,
		}
		if err := attr.create(ctx); err != nil {
			return nil, err
		}
	}
	// Create AttributeValue
	attrValue := AttributeValue{
		AttributeId: attr.Id,
		ProductId:   productId,
		Value:       value,
	}
	if _, err := factory.DB(ctx).Insert(&attrValue); err != nil {
		return nil, err
	}
	return &attrValue, nil
}

func (AttributeValue) CreateOrUpdates(ctx context.Context, productId int64, attrs map[string]string) error {
	attrExtends, err := AttributeExtends{}.getByProductIds(ctx, productId)
	if err != nil {
		return err
	}

	for _, ae := range attrExtends {
		if v, ok := attrs[ae.Name]; ok {
			// update
			if ae.Value != v {
				ae.Value = v
				if err := ae.Update(ctx); err != nil {
					return err
				}
			}
			delete(attrs, ae.Name)
			continue
		}

		// delete
		if err := ae.AttributeValue.Delete(ctx); err != nil {
			return err
		}
	}

	// create
	for k, v := range attrs {
		if _, err := (AttributeValue{}).Create(ctx, productId, k, v); err != nil {
			return err
		}
	}
	return nil
}

func (Attribute) GetByName(ctx context.Context, name string) (*Attribute, error) {
	var a Attribute
	if has, err := factory.DB(ctx).Where("name = ?", name).Get(&a); err != nil {
		return nil, err
	} else if !has {
		return nil, nil
	}
	return &a, nil
}

func (AttributeExtends) getByProductIds(ctx context.Context, productId ...int64) (attrExtends []AttributeExtends, err error) {
	if len(productId) == 0 {
		return
	}
	err = factory.DB(ctx).Table("attribute_value").Select("attribute.*, attribute_value.*").
		Join("INNER", "attribute", "attribute_value.attribute_id = attribute.id").
		In("attribute_value.product_id", productId).Find(&attrExtends)
	return
}

func (av *AttributeValue) Update(ctx context.Context) (err error) {
	_, err = factory.DB(ctx).ID(av.Id).Cols("value").Update(av)
	return
}

func (av *AttributeValue) Delete(ctx context.Context) (err error) {
	_, err = factory.DB(ctx).ID(av.Id).Delete(&AttributeValue{})
	return
}

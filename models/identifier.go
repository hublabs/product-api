package models

import (
	"context"
	"fmt"
	"time"

	"github.com/hublabs/product-api/adapters"
	"github.com/hublabs/product-api/factory"
)

type ProductIdentifier struct {
	Id        int64     `json:"-"`
	ProductId int64     `json:"-" xorm:"index"`
	Uid       string    `json:"uid" xorm:"index"`
	Source    string    `json:"source" xorm:"index"`
	CreatedAt time.Time `json:"-" xorm:"created"`
	UpdatedAt time.Time `json:"-" xorm:"updated"`
}

type SkuIdentifier struct {
	Id        int64     `json:"id,omitempty"`
	SkuId     int64     `json:"skuId,omitempty" xorm:"index"`
	ProductId int64     `json:"productId,omitempty" xorm:"-"`
	Uid       string    `json:"uid" xorm:"index"`
	Source    string    `json:"source,omitempty" xorm:"index"`
	Enable    bool      `json:"enable"`
	CreatedAt time.Time `json:"-" xorm:"created"`
	UpdatedAt time.Time `json:"-" xorm:"updated"`
}

func (ProductIdentifier) GetByUidAndSource(ctx context.Context, uid, source string) (bool, ProductIdentifier, error) {
	p := ProductIdentifier{
		Uid:    uid,
		Source: source,
	}
	exist, err := factory.DB(ctx).Get(&p)
	if err != nil {
		return false, ProductIdentifier{}, err
	}
	if exist {
		return true, p, nil
	}
	return false, ProductIdentifier{}, nil
}

func (SkuIdentifier) GetByUidAndSource(ctx context.Context, uid, source string) (bool, SkuIdentifier, error) {
	s := SkuIdentifier{
		Uid:    uid,
		Source: source,
	}
	exist, err := factory.DB(ctx).Get(&s)
	if err != nil {
		return false, SkuIdentifier{}, err
	}
	if exist {
		return true, s, nil
	}
	return false, SkuIdentifier{}, nil
}

func (p *ProductIdentifier) CreateOrUpdate(ctx context.Context) (err error) {
	var identifier ProductIdentifier
	exist, err := factory.DB(ctx).Where("source = ?", p.Source).And("product_id = ?", p.ProductId).Get(&identifier)
	if err != nil {
		return err
	}
	if !exist || identifier.Id == 0 {
		if err := p.create(ctx); err != nil {
			return err
		}
	} else {
		p.Id = identifier.Id
		if err = p.update(ctx); err != nil {
			return err
		}
	}
	return adapters.MessagePublisher{}.Publish(ctx, *p, adapters.EventProductUidChanged)
}

// Must be private because of event ProductUidChanged
func (s *ProductIdentifier) update(ctx context.Context) (err error) {
	if _, err = factory.DB(ctx).ID(s.Id).Update(s); err != nil {
		return err
	}
	return nil
}

func (s *SkuIdentifier) LoadOrCreate(ctx context.Context) (err error) {
	var identifier SkuIdentifier
	exist, err := factory.DB(ctx).Where("source = ?", s.Source).
		And("uid = ?", s.Uid).
		Get(&identifier)
	if err != nil {
		return err
	}

	if exist {
		if s.SkuId != 0 && s.SkuId != identifier.SkuId {
			return fmt.Errorf("Exist identifier(skuId:%d)", identifier.SkuId)
		}

		*s = identifier
		return nil
	}

	if err := s.create(ctx); err != nil {
		return err
	}

	var sku Sku
	if _, err := factory.DB(ctx).ID(s.SkuId).Get(&sku); err != nil {
		return err
	}
	s.ProductId = sku.ProductId

	return adapters.MessagePublisher{}.Publish(ctx, *s, adapters.EventSkuUidChanged)
}

// Must be private because of event SkuUidChanged
func (s *SkuIdentifier) update(ctx context.Context) (err error) {
	if _, err = factory.DB(ctx).ID(s.Id).Update(s); err != nil {
		return err
	}
	return nil
}

// Must be private because of event ProductUidChanged
func (p *ProductIdentifier) create(ctx context.Context) (err error) {
	_, err = factory.DB(ctx).Insert(p)
	return
}

// Must be private because of event SkuUidChanged
func (s *SkuIdentifier) create(ctx context.Context) (err error) {
	_, err = factory.DB(ctx).Insert(s)
	return
}

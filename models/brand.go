package models

import (
	"context"
	"strconv"

	"github.com/hublabs/product-api/factory"

	"github.com/go-xorm/xorm"
)

type Brand struct {
	Id     int64  `json:"id,omitempty"`
	Code   string `json:"code,omitempty" xorm:"index"`
	Name   string `json:"name,omitempty"`
	Enable bool   `json:"enable" xorm:"index"`
}

func (b *Brand) Create(ctx context.Context) (err error) {
	_, err = factory.DB(ctx).Insert(b)
	return
}

func (b *Brand) Update(ctx context.Context) (err error) {
	_, err = factory.DB(ctx).ID(b.Id).Update(b)
	return
}

func (Brand) GetById(ctx context.Context, id int64) (*Brand, error) {
	b := Brand{Id: id}
	exist, err := factory.DB(ctx).Get(&b)
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, nil
	}
	return &b, nil
}

func (Brand) GetByCode(ctx context.Context, code string) (*Brand, error) {
	b := Brand{Code: code}
	exist, err := factory.DB(ctx).Get(&b)
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, nil
	}
	return &b, nil
}

func (Brand) GetAll(ctx context.Context, q, code, enable string, ids []int64, codes []string, skipCount, maxResultCount int) (int64, []Brand, error) {
	var (
		brands     []Brand
		err        error
		totalCount int64
	)

	query := func() xorm.Interface {
		query := factory.DB(ctx)

		if q != "" {
			query.Where("code LIKE ?", q+"%")
		}

		if code != "" {
			query = query.Where("code = ?", code)
		}

		if len(codes) != 0 {
			query = query.In("code", codes)
		}

		if len(ids) != 0 {
			query = query.In("id", ids)
		}

		if enable != "" {
			b, _ := strconv.ParseBool(enable)
			query.Where("enable = ?", b)
		}

		return query
	}

	if maxResultCount == -1 {
		err = query().Find(&brands)
		totalCount = int64(len(brands))
	} else {
		totalCount, err = query().Limit(maxResultCount, skipCount).FindAndCount(&brands)
	}

	if err != nil {
		return 0, nil, err
	}

	return totalCount, brands, nil
}

func (b *Brand) GetOrCreate(ctx context.Context) error {
	exist, err := factory.DB(ctx).Get(b)
	if err != nil {
		return err
	}
	if !exist {
		_, err = factory.DB(ctx).Insert(b)
		if err != nil {
			return err
		}
	}
	return nil
}

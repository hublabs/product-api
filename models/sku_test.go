package models

import (
	"testing"

	"github.com/pangpanglabs/goutils/test"
)

func TestSkuCRUD(t *testing.T) {
	var id int64
	t.Run("Create", func(t *testing.T) {
		s1 := Sku{
			ProductId: 1,
			Code:      "S001",
			Name:      "sku#1",
			Identifiers: []SkuIdentifier{
				{
					Uid:    "S001001",
					Source: IdentifierSourceBarcode,
				},
			},
		}
		err := s1.Create(ctx)
		test.Ok(t, err)
		test.Equals(t, s1.Id > 0, true)
		id = s1.Id

		s2 := Sku{
			ProductId: 2,
			Code:      "S002",
			Name:      "sku#2",
			Enable:    true,
			Saleable:  true,
		}
		err = s2.Create(ctx)
		test.Ok(t, err)
		test.Equals(t, s2.Id > 0, true)
	})

	t.Run("GetAll", func(t *testing.T) {
		_, count, skus, err := Sku{}.GetAll(ctx, "", "", "", "", "", "", nil, nil, nil, 0, 10, nil, nil, nil, false)
		test.Ok(t, err)
		test.Equals(t, count, int64(2))

		hasMore, _, skus, err := Sku{}.GetAll(ctx, "", "", "", "", "", "", nil, nil, nil, 0, 1, nil, nil, nil, true)
		test.Ok(t, err)
		test.Equals(t, hasMore, true)

		_, count, skus, err = Sku{}.GetAll(ctx, "", "", "S001001", "", "", "", nil, nil, nil, 0, 10, nil, nil, nil, false)
		test.Ok(t, err)
		test.Equals(t, count, int64(1))
		test.Equals(t, skus[0].Id, id)

		_, count, skus, err = Sku{}.GetAll(ctx, "", "", "S001001", "", "true", "", nil, nil, nil, 0, 10, nil, nil, nil, false)
		test.Ok(t, err)
		test.Equals(t, count, int64(0))

		_, count, skus, err = Sku{}.GetAll(ctx, "", "", "", "", "false", "", nil, nil, nil, 0, 10, nil, nil, nil, false)
		test.Ok(t, err)
		test.Equals(t, count, int64(1))

		_, count, skus, err = Sku{}.GetAll(ctx, "", "", "", "", "", "true", nil, nil, nil, 0, 10, nil, nil, nil, false)
		test.Ok(t, err)
		test.Equals(t, count, int64(1))
	})

	t.Run("Get", func(t *testing.T) {
		s, err := Sku{}.GetOne(ctx, id, nil)
		test.Ok(t, err)
		test.Equals(t, s.Id, id)
		test.Equals(t, s.Code, "S001")
		test.Equals(t, s.Name, "sku#1")
	})
}

package models

import (
	"testing"

	"github.com/pangpanglabs/goutils/test"
)

func TestProductCRUD(t *testing.T) {
	var id int64
	var product Product
	t.Run("Create", func(t *testing.T) {
		p1 := Product{
			Code: "P001",
			Name: "product#1",
		}
		rp1, err := p1.CreateOrUpdate(ctx, p1)
		test.Ok(t, err)
		test.Equals(t, rp1.Id > 0, true)
		id = rp1.Id

		p2 := Product{
			Code: "P002",
			Name: "product#2",
		}
		rp2, err := p2.CreateOrUpdate(ctx, p2)
		test.Ok(t, err)
		test.Equals(t, rp2.Id > 0, true)
	})
	t.Run("Get", func(t *testing.T) {
		p, err := Product{}.GetOne(ctx, id, nil)
		test.Ok(t, err)
		test.Equals(t, p.Id, id)
		test.Equals(t, p.Code, "P001")
		test.Equals(t, p.Name, "product#1")
		product = *p
	})
	t.Run("Update", func(t *testing.T) {
		product.Name = "product#1-2"
		_, err := Product{}.CreateOrUpdate(ctx, product)
		test.Ok(t, err)

		c, err := Product{}.GetOne(ctx, product.Id, nil)
		test.Ok(t, err)
		test.Equals(t, c.Name, "product#1-2")
	})
}

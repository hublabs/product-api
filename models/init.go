package models

import (
	"github.com/go-xorm/xorm"
)

func Init(db *xorm.Engine) error {
	if err := db.Sync(new(Product),
		new(Sku),
		new(Option),
		new(ProductIdentifier),
		new(SkuIdentifier),
		new(Price),
		new(Brand),
		new(Attribute),
		new(AttributeValue),
	); err != nil {
		return err
	}
	return nil
}

func DropTables(db *xorm.Engine) error {
	return db.DropTables(new(Product),
		new(Sku),
		new(Option),
		new(ProductIdentifier),
		new(SkuIdentifier),
		new(Price),
		new(Brand),
		new(Attribute),
		new(AttributeValue),
	)
}

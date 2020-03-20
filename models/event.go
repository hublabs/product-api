package models

import (
	"context"
)

const DataSourceContext = "DataSource"

type DataSource string

const (
	DataSourceHandle    DataSource = "handle"
	DataSourceExcel     DataSource = "excel"
	DataSourceInterface DataSource = "interface"
)

func retrieveDataSource(ctx context.Context) DataSource {
	v := ctx.Value(DataSourceContext)
	ds := DataSourceHandle
	if v != nil {
		ds = v.(DataSource)
	}
	return ds
}

type ProductEvent struct {
	Product
	DataSource DataSource `json:"dataSource"`
}

func (p Product) ToEvent(ctx context.Context) interface{} {
	return ProductEvent{
		Product:    p,
		DataSource: retrieveDataSource(ctx),
	}
}

type SkuEvent struct {
	Sku
	DataSource DataSource `json:"dataSource"`
}

func (s Sku) ToEvent(ctx context.Context) interface{} {
	return SkuEvent{
		Sku:        s,
		DataSource: retrieveDataSource(ctx),
	}
}

type PriceEvent struct {
	Price
	DataSource DataSource `json:"dataSource"`
}

func (p Price) ToEvent(ctx context.Context) interface{} {
	return PriceEvent{
		Price:      p,
		DataSource: retrieveDataSource(ctx),
	}
}

type ProductIdentifierEvent struct {
	ProductIdentifier
	DataSource DataSource `json:"dataSource"`
}

func (p ProductIdentifier) ToEvent(ctx context.Context) interface{} {
	return ProductIdentifierEvent{
		ProductIdentifier: p,
		DataSource:        retrieveDataSource(ctx),
	}
}

type SkuIdentifierEvent struct {
	SkuIdentifier
	DataSource DataSource `json:"dataSource"`
}

func (s SkuIdentifier) ToEvent(ctx context.Context) interface{} {
	return SkuIdentifierEvent{
		SkuIdentifier: s,
		DataSource:    retrieveDataSource(ctx),
	}
}

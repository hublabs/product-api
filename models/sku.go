package models

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/hublabs/product-api/adapters"
	"github.com/hublabs/product-api/factory"
)

type Sku struct {
	Id          int64           `json:"id,omitempty"`
	TenantCode  string          `json:"-" xorm:"index varchar(16)"`
	ProductId   int64           `json:"productId,omitempty" xorm:"index"`
	Code        string          `json:"code" xorm:"index varchar(64)"`
	Name        string          `json:"name,omitempty"`
	Image       string          `json:"image,omitempty"`
	Identifiers []SkuIdentifier `json:"identifiers,omitempty" xorm:"-"`
	Options     []Option        `json:"options,omitempty" xorm:"-"`
	Product     *Product        `json:"product,omitempty" xorm:"-"`
	Enable      bool            `json:"enable" xorm:"index"`
	Saleable    bool            `json:"saleable" xorm:"index"`
	CreatedAt   time.Time       `json:"createdAt,omitempty" xorm:"created"`
	UpdatedAt   time.Time       `json:"updatedAt,omitempty" xorm:"updated"`
	DeletedAt   time.Time       `json:"-" xorm:"deleted index"`
}

func (Sku) GetSimple(ctx context.Context, skuId int64, fields FieldTypeList) (*Sku, error) {
	var rows []struct {
		Sku     Sku     `xorm:"extends"`
		Product Product `xorm:"extends"`
		Brand   Brand   `xorm:"extends"`
		Option  Option  `xorm:"extends"`
	}
	if err := factory.DB(ctx).Table("sku").
		Join("INNER", "product", "sku.product_id = product.id").
		Join("INNER", "brand", "product.brand_id = brand.id").
		Join("LEFT", "`option`", "sku.id = `option`.sku_id").
		Where("sku.id = ?", skuId).
		Where(excludeDeleted("product")).
		Find(&rows); err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}

	sku := rows[0].Sku
	sku.Product = &rows[0].Product
	sku.Product.Brand = rows[0].Brand

	if fields.Contains(FieldTypeAttribute) {
		if err := sku.Product.LoadAttributes(ctx); err != nil {
			return nil, err
		}
	}

	for _, row := range rows {
		if row.Option.Id != 0 {
			sku.Options = append(sku.Options, row.Option)
		}
	}
	return &sku, nil
}

func (Sku) GetOne(ctx context.Context, id int64, fields FieldTypeList) (*Sku, error) {
	var skus SkuList
	if err := factory.DB(ctx).Where("id = ?", id).Limit(1).Find(&skus); err != nil {
		return nil, err
	} else if len(skus) == 0 {
		return nil, nil
	}

	if err := skus.LoadProducts(ctx, fields); err != nil {
		return nil, err
	}

	if err := skus.LoadIdentifiers(ctx); err != nil {
		return nil, err
	}

	if err := skus.LoadOptions(ctx); err != nil {
		return nil, err
	}

	return &skus[0], nil
}

func (Sku) GetAll(ctx context.Context, q, productCode, barcode, brandCode, enable, saleable string, codes []string, ids, brandIds []int64, skipCount, maxResultCount int, sortby, order []string, fields FieldTypeList, withHasMore bool) (bool, int64, []Sku, error) {
	query := fmt.Sprintf(`SELECT sku.* FROM sku
INNER JOIN product ON sku.product_id = product.id
WHERE sku.tenant_code = ? AND (%s)`, excludeDeleted("sku"))
	args := []interface{}{tenantCode(ctx)}
	if len(ids) != 0 {
		placeholder := strings.Repeat("?,", len(ids))
		query = fmt.Sprintf(query+" AND sku.id IN (%s)", placeholder[:len(placeholder)-1])
		for _, id := range ids {
			args = append(args, id)
		}
	}
	if len(brandIds) != 0 {
		placeholder := strings.Repeat("?,", len(brandIds))
		query = fmt.Sprintf(query+" AND product.brand_id IN (%s)", placeholder[:len(placeholder)-1])
		for _, id := range brandIds {
			args = append(args, id)
		}
	}
	if brandCode != "" {
		query = query + " AND product.brand_id IN (SELECT id FROM brand WHERE code = ?)"
		args = append(args, brandCode)
	}
	if productCode != "" {
		query = query + " AND product.code = ?"
		args = append(args, productCode)
	}
	if len(codes) != 0 {
		placeholder := strings.Repeat("?,", len(codes))
		query = fmt.Sprintf(query+" AND sku.code IN (%s)", placeholder[:len(placeholder)-1])
		for _, code := range codes {
			args = append(args, code)
		}
	}
	if barcode != "" {
		query = query + " AND sku.id IN (SELECT sku_id FROM sku_identifier WHERE uid = ? AND source = ?)"
		args = append(args, barcode, IdentifierSourceBarcode)
	}
	if enable != "" {
		b, _ := strconv.ParseBool(enable)
		query = query + " AND sku.enable = ?"
		args = append(args, b)
	}
	if saleable != "" {
		b, _ := strconv.ParseBool(saleable)
		query = query + " AND sku.saleable = ?"
		args = append(args, b)
	}
	if q != "" {
		query = query + " AND sku.code LIKE ?\nUNION\n" + query + " AND sku.id IN (SELECT sku_id FROM sku_identifier WHERE uid LIKE ?)"
		args = append(args, q+"%")
		args = append(args, args...)
	}

	var (
		hasMore    bool
		totalCount int64
		err        error
	)
	if !withHasMore {
		totalCount, err = factory.DB(ctx).SQL(fmt.Sprintf("SELECT COUNT(*) FROM (\n%s\n) T", query), args...).Count(&Sku{})
		if err != nil {
			return false, 0, nil, err
		}
	}

	orderQuery := "ORDER BY "
	if len(sortby) == 0 || len(order) == 0 {
		sortby = []string{"id"}
		order = []string{"desc"}
	}
	for i, s := range sortby {
		if len(order) <= i {
			break
		}
		order[i] = strings.ToUpper(order[i])
		if order[i] != "DESC" && order[i] != "ASC" {
			break
		}
		if i > 0 {
			orderQuery += ", "
		}
		orderQuery += fmt.Sprintf("`%s` %s", s, order[i]) // TODO: Anti SQL injection
	}
	limitQuery := fmt.Sprintf("LIMIT %d", maxResultCount)
	if withHasMore {
		limitQuery = fmt.Sprintf("LIMIT %d", maxResultCount+1)
	}
	if skipCount > 0 {
		limitQuery = limitQuery + fmt.Sprintf(" OFFSET %d", skipCount)
	}
	var skus SkuList
	if err := factory.DB(ctx).SQL(fmt.Sprintf("SELECT * FROM (\n%s\n) T %s %s", query, orderQuery, limitQuery), args...).Find(&skus); err != nil {
		return false, 0, nil, err
	}
	if withHasMore && len(skus) == maxResultCount+1 {
		skus = skus[:maxResultCount]
		hasMore = true
	}

	if len(skus) == 0 {
		return false, 0, nil, nil
	}

	if err := skus.LoadProducts(ctx, fields); err != nil {
		return false, 0, nil, err
	}

	if err := skus.LoadIdentifiers(ctx); err != nil {
		return false, 0, nil, err
	}

	if err := skus.LoadOptions(ctx); err != nil {
		return false, 0, nil, err
	}

	return hasMore, totalCount, skus, nil
}

func (Sku) GetByProductId(ctx context.Context, id int64) ([]Sku, error) {
	var skus []Sku
	if err := factory.DB(ctx).Table("sku").
		Where("product_id = ?", id).
		Desc("id").
		Find(&skus); err != nil {
		return nil, err
	}
	return skus, nil
}

func (s *Sku) Update(ctx context.Context) (err error) {
	cols := []string{
		"code", "name", "image",
	}
	if _, err = factory.DB(ctx).ID(s.Id).Cols(cols...).Update(s); err != nil {
		return
	}
	var options []Option
	if err := factory.DB(ctx).Where("sku_id = ?", s.Id).Find(&options); err != nil {
		return err
	}
OptionLoop:
	for i := range s.Options {
		s.Options[i].SkuId = s.Id
		for j := range options {
			if s.Options[i].Id == options[j].Id {
				if err := s.Options[i].update(ctx); err != nil {
					return err
				}
				continue OptionLoop
			}
		}
		if err := s.Options[i].create(ctx); err != nil {
			return err
		}
	}

	for i := range s.Identifiers {
		s.Identifiers[i].SkuId = s.Id
		if s.Identifiers[i].Id != 0 {
			if err := s.Identifiers[i].update(ctx); err != nil {
				return err
			}
		} else {
			if err = s.Identifiers[i].LoadOrCreate(ctx); err != nil {
				return err
			}
		}
	}

	if err := s.removeOptionsExcept(ctx, s.Options); err != nil {
		return err
	}

	if err := s.removeIdentifiersExcept(ctx, s.Identifiers); err != nil {
		return err
	}

	return adapters.MessagePublisher{}.Publish(ctx, *s, adapters.EventSkuChanged)
}

func (s *Sku) Create(ctx context.Context) error {
	s.TenantCode = tenantCode(ctx)
	if _, err := factory.DB(ctx).Insert(s); err != nil {
		return err
	}
	for i := range s.Identifiers {
		s.Identifiers[i].SkuId = s.Id
		var d SkuIdentifier
		exist, err := factory.DB(ctx).Where("uid = ?", s.Identifiers[i].Uid).And("source = ?", s.Identifiers[i].Source).Get(&d)
		if err != nil {
			return err
		}
		if !exist {
			if err := s.Identifiers[i].create(ctx); err != nil {
				return err
			}
		} else {
			if err := s.Identifiers[i].update(ctx); err != nil {
				return err
			}
		}

	}
	for j := range s.Options {
		s.Options[j].SkuId = s.Id
		if err := s.Options[j].create(ctx); err != nil {
			return err
		}
	}
	return adapters.MessagePublisher{}.Publish(ctx, *s, adapters.EventSkuAdded)
}

func (s Sku) removeOptionsExcept(ctx context.Context, except []Option) (err error) {
	var exceptOptionIds []int64
	for _, option := range except {
		exceptOptionIds = append(exceptOptionIds, option.Id)
	}

	if len(exceptOptionIds) > 0 {
		if _, err := factory.DB(ctx).Where("sku_id = ?", s.Id).NotIn("id", exceptOptionIds).Delete(&Option{}); err != nil {
			return err
		}
	}
	return nil
}

func (s Sku) removeIdentifiersExcept(ctx context.Context, except []SkuIdentifier) (err error) {
	var exceptIdentifilerIds []int64
	for _, identifier := range except {
		exceptIdentifilerIds = append(exceptIdentifilerIds, identifier.Id)
	}
	_, err = factory.DB(ctx).Where("sku_id = ?", s.Id).NotIn("id", exceptIdentifilerIds).Delete(&SkuIdentifier{})
	return
}

func (Sku) SearchAll(ctx context.Context, q, enable, saleable string, filter Filter, skipCount, maxResultCount int, sortby, order []string, fields FieldTypeList, withHasMore bool) (bool, int64, []Sku, error) {
	query := factory.DB(ctx).Table("sku").Select("sku.*").Join("INNER", "product", "product.id = sku.product_id").
		Where("sku.tenant_code = ?", tenantCode(ctx))
	if q != "" {
		query.Where(`sku.id IN ( SELECT id FROM (
    SELECT id FROM sku WHERE code LIKE ?
    UNION
    SELECT sku.id FROM sku JOIN sku_identifier ON sku.id = sku_identifier.sku_id WHERE sku_identifier.uid LIKE ?
) T )`, q+"%", q+"%")
	}
	if enable != "" {
		b, _ := strconv.ParseBool(enable)
		query.Where("sku.enable = ?", b)
	}
	if saleable != "" {
		b, _ := strconv.ParseBool(saleable)
		query = query.Where("sku.saleable = ?", b)
	}
	filterQuery(query, filter)

	var (
		skus       SkuList
		hasMore    bool
		totalCount int64
		err        error
	)

	if len(sortby) == 0 || len(order) == 0 {
		sortby = []string{"sku.id"}
		order = []string{"desc"}
	}

	if err = setSortOrder(query, sortby, order); err != nil {
		return false, 0, nil, err
	}

	if withHasMore {
		err = query.Limit(maxResultCount+1, skipCount).Find(&skus)
		if len(skus) == maxResultCount+1 {
			skus = skus[:maxResultCount]
			hasMore = true
		}
	} else {
		totalCount, err = query.Limit(maxResultCount, skipCount).FindAndCount(&skus)
	}
	if err != nil {
		return false, 0, nil, err
	}

	if len(skus) == 0 {
		return false, 0, nil, nil
	}

	if err := skus.LoadProducts(ctx, fields); err != nil {
		return false, 0, nil, err
	}

	if err := skus.LoadIdentifiers(ctx); err != nil {
		return false, 0, nil, err
	}

	if err := skus.LoadOptions(ctx); err != nil {
		return false, 0, nil, err
	}

	return hasMore, totalCount, skus, nil
}

func (Sku) GetByUids(ctx context.Context, source string, fields FieldTypeList, uids ...string) ([]Sku, error) {
	var skuIds []int64
	query := factory.DB(ctx).Table("sku_identifier").Select("sku_id").Distinct("sku_id").In("uid", uids)
	if source != "" {
		query.Where("source = ?", source)
	}
	if err := query.Find(&skuIds); err != nil {
		return nil, err
	}
	if len(skuIds) == 0 {
		return nil, nil
	}

	var skus SkuList
	if err := factory.DB(ctx).Table("sku").
		Where("tenant_code = ?", tenantCode(ctx)).
		In("id", skuIds).
		Find(&skus); err != nil {
		return nil, err
	}

	if err := skus.LoadProducts(ctx, fields); err != nil {
		return nil, err
	}

	if err := skus.LoadIdentifiers(ctx); err != nil {
		return nil, err
	}

	if err := skus.LoadOptions(ctx); err != nil {
		return nil, err
	}

	return skus, nil
}

type SkuList []Sku

func (skus SkuList) ProductIds() (ids []interface{}) {
	m := map[int64]bool{}
	for _, s := range skus {
		if _, exist := m[s.ProductId]; !exist {
			ids = append(ids, s.ProductId)
			m[s.ProductId] = true
		}
	}
	return
}

func (skus SkuList) Ids() (ids []interface{}) {
	for _, s := range skus {
		ids = append(ids, s.Id)
	}
	return
}

func (skus SkuList) Find(id int64) *Sku {
	for i := range skus {
		if skus[i].Id == id {
			return &skus[i]
		}
	}
	return nil
}

func (skus SkuList) FindByProductId(productId int64) []*Sku {
	var result []*Sku
	for i := range skus {
		if skus[i].ProductId == productId {
			result = append(result, &skus[i])
		}
	}
	return result
}

func (skus SkuList) LoadOptions(ctx context.Context) error {
	var options []Option
	if err := factory.DB(ctx).In("sku_id", skus.Ids()...).Find(&options); err != nil {
		return err
	}
	for _, option := range options {
		s := skus.Find(option.SkuId)
		if s != nil {
			s.Options = append(s.Options, option)
		}
	}
	return nil
}

func (skus SkuList) LoadProducts(ctx context.Context, fields FieldTypeList) error {
	var products ProductList
	if err := factory.DB(ctx).In("id", skus.ProductIds()...).Find(&products); err != nil {
		return err
	}

	if err := products.LoadPrices(ctx); err != nil {
		return err
	}

	if err := products.LoadBrands(ctx); err != nil {
		return err
	}

	if fields.Contains(FieldTypeAttribute) {
		if err := products.LoadAttributes(ctx); err != nil {
			return err
		}
	}

	for _, product := range products {
		for _, s := range skus.FindByProductId(product.Id) {
			p := product
			s.Product = &p
		}
	}

	return nil
}

func (skus SkuList) LoadIdentifiers(ctx context.Context) error {
	var identifiers []SkuIdentifier
	if err := factory.DB(ctx).In("sku_id", skus.Ids()...).Find(&identifiers); err != nil {
		return err
	}
	for _, identifier := range identifiers {
		s := skus.Find(identifier.SkuId)
		if s != nil {
			s.Identifiers = append(s.Identifiers, identifier)
		}
	}
	return nil
}

func (t *Sku) FromDB(data []byte) error {
	if err := json.Unmarshal(data, t); err != nil {
		log.Println("Json Unmarshal Error(Sku). json:", string(data))
	}
	return nil
}

func (t Sku) ToDB() ([]byte, error) {
	return json.Marshal(t)
}

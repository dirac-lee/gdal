package gdal

import (
	"context"
)

type DAL struct {
	Client *client
	MakePO func() any
}

// Create 创建
func (r *DAL) Create(ctx context.Context, po any) error {
	return r.Client.Create(ctx, po)
}

// Upsert 保存或者创建, 注意：po 必须有 ID 字段
func (r *DAL) Upsert(ctx context.Context, po, where, update any) (isCreated bool, err error) {
	return r.Client.Upsert(ctx, po, where, update)
}

// Delete 删除
func (r *DAL) Delete(ctx context.Context, where any) (int64, error) {
	return r.Client.Delete(ctx, r.MakePO(), where)
}

// DeleteByID 通过 id 删除
func (r *DAL) DeleteByID(ctx context.Context, id int64) (int64, error) {
	return r.Client.Delete(ctx, r.MakePO(), idWhere{ID: &id})
}

// UpdateByID 通过 id 更新
func (r *DAL) UpdateByID(ctx context.Context, id int64, update any) (int64, error) {
	return r.Client.Update(ctx, r.MakePO(), idWhere{ID: &id}, update)
}

// Update 更新
func (r *DAL) Update(ctx context.Context, where any, update any) (int64, error) {
	return r.Client.Update(ctx, r.MakePO(), where, update)
}

// Count 计数
func (r *DAL) Count(ctx context.Context, where any) (int32, error) {
	return r.Client.Count(ctx, r.MakePO(), where)
}

// Find 查找
func (r *DAL) Find(ctx context.Context, pos any, where any, options ...QueryOption) error {
	return r.Client.Find(ctx, pos, where, options...)
}

// Exist 查找是否存在
func (r *DAL) Exist(ctx context.Context, where any) (bool, error) {
	return r.Client.Exist(ctx, r.MakePO(), where)
}

// First 首条记录
func (r *DAL) First(ctx context.Context, po, where any, options ...QueryOption) error {
	return r.Client.First(ctx, po, where, options...)
}

type QueryConfig struct {
	readMaster bool
	debug      bool

	// export field
	Limit   *int
	Offset  *int
	Order   *string
	Selects []string
}

type QueryOption func(v *QueryConfig)

// MakeQueryConfig
// @Description: 将 options 转化为 QueryConfig
// @param options:
// @return *QueryConfig:
func MakeQueryConfig(options []QueryOption) *QueryConfig {
	opt := new(QueryConfig)
	for _, v := range options {
		if v != nil {
			v(opt)
		}
	}
	return opt
}

// WithMaster
// @Description: 查主库
// @return QueryOption:
func WithMaster() QueryOption {
	return func(v *QueryConfig) {
		v.readMaster = true
	}
}

// WithLimit
// @Description: 指定限制条数
// @param limit: 限制条数
// @return QueryOption:
func WithLimit(limit int) QueryOption {
	return func(v *QueryConfig) {
		v.Limit = &limit
	}
}

// WithOffset
// @Description: 指定查询偏移
// @param offset: 查询偏移
// @return QueryOption:
func WithOffset(offset int) QueryOption {
	return func(v *QueryConfig) {
		v.Offset = &offset
	}
}

// WithOrder
// @Description: 指定查询顺序
// @param order: 查询顺序
// @return QueryOption:
func WithOrder(order string) QueryOption {
	return func(v *QueryConfig) {
		v.Order = &order
	}
}

// WithSelects
// @Description: 指定查询字段
// @param selects: 查询字段
// @return QueryOption:
func WithSelects(selects []string) QueryOption {
	return func(v *QueryConfig) {
		v.Selects = selects
	}
}

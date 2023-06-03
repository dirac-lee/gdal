package gdal

import (
	"context"
	"github.com/dirac-lee/gdal"
	"github.com/dirac-lee/gdal/generic/gptr"
	"github.com/dirac-lee/gdal/generic/gslice"
	"github.com/dirac-lee/gdal/generic/gvalue"
	"github.com/dirac-lee/gdal/gerror"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// GenericDAL
// @Description: 泛型 DAL，提供通用增删改查功能。
// 业务 DAL 嵌套 *GenericDAL，并指定 PO、Where 和 Update 结构体，然后进行针对性扩展。
// @param PO     持久对象 (persistent object)
// @param Where  查询条件对象 (query condition object)
// @param Update 更新规则对象 (update rule object)
type GenericDAL[PO schema.Tabler, Where any, Update any] struct {
	*gdal.DAL
}

func NewGenericDAL[PO schema.Tabler, Where any, Update any](cli gdal.Client) *GenericDAL[PO, Where, Update] {
	return &GenericDAL[PO, Where, Update]{
		cli.DAL(func() any {
			return gvalue.Zero[PO]()
		}),
	}
}

// TableName
// @Description: 对应表名
// @return string:
func (dal *GenericDAL[PO, Where, Update]) TableName() string {
	return gvalue.Zero[PO]().TableName()
}

// Create
// @Description: 插入单条记录
// @param ctx:
// @param po:
// @return error:
func (dal *GenericDAL[PO, Where, Update]) Create(ctx context.Context, po *PO) error {
	return dal.DAL.Create(ctx, po)
}

// MCreate
// @Description: 插入多条记录
// @param ctx:
// @param pos: 多条记录，由于需要回写 ID，所以使用指针
// @return int64: 成功插入个数
// @return error:
func (dal *GenericDAL[PO, Where, Update]) MCreate(ctx context.Context, pos *[]*PO) (int64, error) {
	tx := dal.DAL.Client.DB(ctx).Table(dal.TableName()).CreateInBatches(pos, 100)
	return tx.RowsAffected, tx.Error
}

// Count
// @Description: 根据查询条件返回记录个数
// @param ctx:
// @param where: 查询条件
// @return int64: 记录个数
// @return error:
func (dal *GenericDAL[PO, Where, Update]) Count(ctx context.Context, where *Where) (int64, error) {
	injectDefaultIfHas(where)                     // 如果配置了默认值，并且用户未指定该字段，则注入默认值
	indexedDAL := dal.forceIndexIfHas(ctx, where) // 如果 Where 指定了强制索引，则走强制索引
	count, err := indexedDAL.DAL.Count(ctx, where)
	return int64(count), err
}

// Find
// @Description:  条件查询
// @param ctx:
// @param pos: 持久(子)对象列表指针，与数据库交互的字段需要包含 `gorm:"column:列名"` tag
// @param where: 查询条件
// @param options: 分页规则
// @return error:
func (dal *GenericDAL[PO, Where, Update]) Find(ctx context.Context, pos any, where any, options ...gdal.QueryOption) error {
	selector := GetSelectorFromPOs(pos)                             // 根据 PO gorm tag 确定 select 字段列表
	injectDefaultIfHas(where)                                       // 如果配置了默认值，并且用户未指定该字段，则注入默认值
	indexedDAL := dal.forceIndexIfHas(ctx, where)                   // 如果 Where 指定了强制索引，则走强制索引
	options = gslice.Insert(options, 0, gdal.WithSelects(selector)) // 优先使用业务指定的 WithSelects
	err := indexedDAL.DAL.Find(ctx, pos, where, options...)
	if gerror.IsRecordNotFoundError(err) {
		return nil
	}
	return err
}

// First
// @Description:  条件查询单个记录
// @param ctx:
// @param po: 结果指针
// @param where: 查询条件
// @param options: 分页规则
// @return error:
func (dal *GenericDAL[PO, Where, Update]) First(ctx context.Context, po any, where any, options ...gdal.QueryOption) error {
	selector := GetSelectorFromPOs(po)                              // 根据 PO gorm tag 确定 select 字段列表
	injectDefaultIfHas(where)                                       // 如果配置了默认值，并且用户未指定该字段，则注入默认值
	indexedDAL := dal.forceIndexIfHas(ctx, where)                   // 如果 Where 指定了强制索引，则走强制索引
	options = gslice.Insert(options, 0, gdal.WithSelects(selector)) // 优先使用业务指定的 WithSelects
	return indexedDAL.DAL.First(ctx, po, where, options...)
}

// MQuery
// @Description: 根据查询条件返回所有记录 (引用 Find)
// @param ctx:
// @param where: 查询条件
// @return []*PO: 所有记录
// @return error:
func (dal *GenericDAL[PO, Where, Update]) MQuery(ctx context.Context, where *Where, options ...gdal.QueryOption) ([]*PO, error) {
	var pos []*PO
	err := dal.Find(ctx, &pos, where, options...)
	return pos, err
}

// MQueryByIDs
// @Description: 根据主键ID列表返回记录
// @param ctx:
// @param ids: 主键ID列表
// @return []*PO:
// @return error:
func (dal *GenericDAL[PO, Where, Update]) MQueryByIDs(ctx context.Context, ids []int64, options ...gdal.QueryOption) ([]*PO, error) {
	where := &idWhere{
		IDIn: ids,
	}
	var pos []*PO
	err := dal.Find(ctx, &pos, where, options...)
	return pos, err
}

// MQueryByPagingOpt
// @Description: 根据查询条件分页查询 (引用 Count 和 Find)
// @param ctx:
// @param where: 查询条件
// @param options: 分页配置
// @return []*PO: 本页记录
// @return int64: 满足查询条件的记录总数
// @return error:
func (dal *GenericDAL[PO, Where, Update]) MQueryByPagingOpt(ctx context.Context, where *Where, options ...gdal.QueryOption) ([]*PO, int64, error) {
	count, err := dal.Count(ctx, where)
	if err != nil || count == 0 {
		return nil, 0, err
	}

	var pos []*PO
	pos, err = dal.MQuery(ctx, where, options...)
	return pos, count, err
}

// MQueryByPaging
// @Description: 根据查询条件分页查询 (引用 Count 和 Find)
// @param ctx:
// @param where: 查询条件
// @param limit: 页大小限制
// @param offset: 偏移量
// @param order: 排序规则
// @return []*PO: 本页记录
// @return int64: 满足查询条件的记录总数
// @return error:
func (dal *GenericDAL[PO, Where, Update]) MQueryByPaging(ctx context.Context, where *Where, limit *int64, offset *int64, order *string) ([]*PO, int64, error) {
	options := buildQueryOptions(limit, offset, order)
	return dal.MQueryByPagingOpt(ctx, where, options...)
}

// QueryFirst
// @Description: 根据查询条件返回第一条记录
// @param ctx:
// @param where: 查询条件
// @return *PO: 第一条记录
// @return error:
func (dal *GenericDAL[PO, Where, Update]) QueryFirst(ctx context.Context, where *Where) (*PO, error) {
	var po PO
	err := dal.First(ctx, &po, where)
	if err != nil {
		if gerror.IsRecordNotFoundError(err) {
			return nil, nil
		}
		return nil, err
	}
	return &po, nil
}

// QueryByID
// @Description: 根据主键ID返回唯一一条记录
// @param ctx:
// @param id: 主键ID
// @return *PO: 唯一一条记录
// @return error:
func (dal *GenericDAL[PO, Where, Update]) QueryByID(ctx context.Context, id int64) (*PO, error) {
	where := &idWhere{
		ID: gptr.Of(id),
	}
	var po PO
	err := dal.First(ctx, &po, where)
	if err != nil {
		if gerror.IsRecordNotFoundError(err) {
			return nil, nil
		}
		return nil, err
	}
	return &po, err
}

// MUpdate
// @Description: 根据查询条件进行更新
// @param ctx:
// @param where: 查询条件
// @param update: 更新值
// @return int64: 被更新记录的个数
// @return error:
func (dal *GenericDAL[PO, Where, Update]) MUpdate(ctx context.Context, where *Where, update *Update) (int64, error) {
	injectDefaultIfHas(where) // 如果配置了默认值，并且用户未指定该字段，则注入默认值
	return dal.DAL.Update(ctx, where, update)
}

// Update
// @Description: 根据查询条件进行更新
// @param ctx:
// @param where: 查询条件
// @param update: 更新值
// @return error:
func (dal *GenericDAL[PO, Where, Update]) Update(ctx context.Context, where *Where, update *Update) error {
	injectDefaultIfHas(where) // 如果配置了默认值，并且用户未指定该字段，则注入默认值
	_, err := dal.MUpdate(ctx, where, update)
	return err
}

// UpdateByID
// @Description: 根据主键ID进行更新
// @param ctx:
// @param id: 主键ID
// @param update: 更新值
// @return error:
func (dal *GenericDAL[PO, Where, Update]) UpdateByID(ctx context.Context, id int64, update *Update) error {
	_, err := dal.DAL.UpdateByID(ctx, id, update)
	return err
}

// Delete
// @Description: ❗️物理删除，逻辑删除请业务 DAL override
// @param ctx:
// @param where: 查询条件
// @return int64: 删除条数
// @return error:
func (dal *GenericDAL[PO, Where, Update]) Delete(ctx context.Context, where *Where) (int64, error) {
	return dal.DAL.Delete(ctx, where)
}

// DeleteByID
// @Description: ❗️物理删除，逻辑删除请业务 DAL override
// @param ctx:
// @param id: 主键ID
// @return int64: 删除条数
// @return error:
func (dal *GenericDAL[PO, Where, Update]) DeleteByID(ctx context.Context, id int64) (int64, error) {
	return dal.DAL.DeleteByID(ctx, id)
}

// WithTx
// @Description: 生成事务 DAL，使用同一 tx 构造的事务 DAL 引用同一 tx，所以能够支持事务
// @param tx: 事务 db 对象
// @return *GenericDAL: 事务 DAL
func (dal *GenericDAL[PO, Where, Update]) WithTx(tx *gorm.DB) *GenericDAL[PO, Where, Update] {
	return NewGenericDAL[PO, Where, Update](gdal.New(tx))
}

// DB
// @Description: 返回当前 DAL 引用的 db 对象，以支持使用原生 gorm 生成更复杂的 sql
// @param ctx: 自动 WithCtx
// @param options: db 配置，如配置主键、本地缓存、读主库等
// @return *gorm.db: 当前 DAL 引用的 db 对象
func (dal *GenericDAL[PO, Where, Update]) DB(ctx context.Context, options ...gdal.QueryOption) *gorm.DB {
	return dal.Client.DB(ctx, options...)
}

func buildQueryOptions(limit *int64, offset *int64, order *string) []gdal.QueryOption {
	var options []gdal.QueryOption
	if limit != nil {
		options = append(options, gdal.WithLimit(int(*limit)))
	}
	if offset != nil {
		options = append(options, gdal.WithOffset(int(*offset)))
	}
	if order != nil {
		options = append(options, gdal.WithOrder(*order))
	}
	return options
}
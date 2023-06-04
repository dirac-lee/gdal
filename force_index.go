package gdal

import (
	"context"
	"github.com/dirac-lee/gdal/gutil/greflect"
	"gorm.io/hints"
)

// ForceIndexer
// @Description: Where 指定的强制索引。不实现或返回空串则不指定。
type ForceIndexer interface {
	ForceIndex() string
}

// forceIndexIfHas
// @Description: 如果 Where 实现了 ForceIndexer 接口，走 ForceIndexer 指定的索引，否则由数据库自选
// ❗请为️ Where 而不是 *Where 实现此接口，因为这里使用 Where 来判断是否实现该接口
// @param ctx:
// @return *GDAL
func (gdal *GDAL[PO, Where, Update]) forceIndexIfHas(ctx context.Context, where any) *GDAL[PO, Where, Update] {
	txDAL := gdal
	var forceIndex string
	if forceIndexer, implemented := greflect.Implements[ForceIndexer](where); implemented { // Where 指定了全局索引
		forceIndex = forceIndexer.ForceIndex()
	}
	if len(forceIndex) == 0 { // Where 没有指定强制索引，由数据库自行决定
		return txDAL
	}
	return NewGDAL[PO, Where, Update](gdal.DBWithCtx(ctx).Clauses(hints.UseIndex(forceIndex)))
}

package gdal

import (
	"context"

	"github.com/dirac-lee/gdal/gutil/greflect"
	"gorm.io/hints"
)

// ForceIndexer assign force index for Where. no force index when return "" or unimplemented.
//
// ⚠️  WARNING: please implement ForceIndexer for Where in spite of *Where
type ForceIndexer interface {
	ForceIndex() string
}

// forceIndexIfHas  if Where implements ForceIndexer，force the index by ForceIndexer, otherwise, it depends on db
//
// 💡 HINT:
//
// ⚠️  WARNING: please implement ForceIndexer for Where in spite of *Where
//
// 🚀 example:
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

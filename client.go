package gdal

import (
	"context"
	"gorm.io/gorm"
)

type Client interface {
	Create(ctx context.Context, po any) error
	Delete(ctx context.Context, po any, where any) (int64, error)
	Update(ctx context.Context, po any, where any, update any) (int64, error)
	Find(ctx context.Context, po any, where any, options ...QueryOption) (err error)
	First(ctx context.Context, po, where any, options ...QueryOption) error
	Count(ctx context.Context, po any, where any) (int32, error)
	Exist(ctx context.Context, po any, where any) (bool, error)
	Save(ctx context.Context, po any) (int64, error)
	DB(ctx context.Context, options ...QueryOption) *gorm.DB
	DAL(makePO func() any) *DAL
}

// client gdal 实例
type client struct {
	db *gorm.DB
}

func New(db *gorm.DB, options ...ClientOption) Client {
	cli := &client{
		db: db,
	}
	for _, option := range options {
		option(cli)
	}
	return cli
}

type ClientOption func(x *client)

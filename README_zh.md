# GDAL

泛型数据访问层。基于 gorm 实现，提供一套方便的 ORM 语法。

[English](./README.md)

## 1. 安装

> 前置条件：go > 1.18

```bash
go get github.com/dirac-lee/gdal
```

## 2. 使用方式

具体可参考 [example](./example)

### 2.1 定义三种业务结构体

### 2.1.1 三种业务结构体

- 数据行模型结构体 PO：结构体类型，通过 `gorm` tag 映射为数据库表中的一行
- 数据行筛选结构体 Where：结构体类型，通过 `sql_field` tag 和 `sql_operator` tag 映射为 where 条件
- 数据行更改结构体 Update：结构体类型，通过 `sql_field` tag 和 `sql_expr` tag 映射为 update 规则

> ⚠️ 注意：「数据行筛选结构体」与「数据行更改结构体」均要求所有字段都是指针，可以使用 [gptr](./gutil/gptr/gptr.go) 来简化。

### 2.1.2 三种业务结构体示例

```go
// Package model
// @Author liming.dirac
// @Date 2023/6/4
// @Description:
package model

import "time"

// User db model struct, will be mapped to row of db by ORM.
type User struct {
	ID         int64     `gorm:"column:id"`
	Name       string    `gorm:"column:name"`
	Balance    int64     `gorm:"column:balance"`
	CreateTime time.Time `gorm:"column:create_time"`
	UpdateTime time.Time `gorm:"column:update_time"`
	Deleted    bool      `gorm:"column:deleted"`
}

// TableName the corresponding table name of User model struct
func (po User) TableName() string {
	return "user"
}

// UserWhere db where struct, will be mapped to SQL where condition by ORM.
type UserWhere struct {
	ID           *int64     `sql_field:"id" sql_operator:"="`
	IDIn         []int64    `sql_field:"id" sql_operator:"in"`
	Name         *string    `sql_field:"name" sql_operator:"="`
	NameLike     *string    `sql_field:"name" sql_operator:"like"`
	CreateTimeGT *time.Time `sql_field:"create_time" sql_operator:">"`
	Deleted      *bool      `sql_field:"deleted" sql_operator:"="`
}

type UserUpdate struct {
	ID           *int64     `sql_field:"id"`
	Name         *string    `sql_field:"name"`
	Balance      *int64     `gorm:"column:balance"`
	BalanceAdd   *int64     `gorm:"column:balance" sql_expr:"+"`
	BalanceMinus *int64     `gorm:"column:balance" sql_expr:"-"`
	UpdateTime   *time.Time `sql_field:"update_time"`
	Deleted      *bool      `sql_field:"deleted"`
}
```

#### 2.1.3 数据行筛选结构体 Where 特性

- 使用 `sql_field` tag 来指示映射列
    - `sql_field:"id"`， 表示 `where id = ?`
    - `sql_field:"dog.id"`， 表示 `where dog.id = ?`
- 使用 `sql_operator` tag 来指示操作符
    - `sql_operator:"="`，表示：`where name = ?`，默认
    - `sql_operator:">"`，表示：`where age > ? `
    - `sql_operator:">="`，表示：`where age >= ? `
    - `sql_operator:"<"`，表示：`where age < ? `
    - `sql_operator:"<="`，表示：`where age <= ? `
    - `sql_operator:"!="`，表示：`where age != ? `
    - `sql_operator:"null"`，类型必须是 `*bool`
        - 当为 `true`，表示：`where age is null`，
        - 当为 `false`，表示：`where age is not null`
    - `sql_operator:"in"`，表示：`where id in (?)`
      > ⚠️ 注意：字段若是空slice([]int) 条件将被忽略；可使用非空slice指针(*[]int)避免
    - `sql_operator:"not in"`，表示：`where id not in (?)`，
      > ⚠️ 注意：同 in
    - `sql_operator:"like"`，表示：`(where name like ?, name)`
    - `sql_operator:"left like"`，表示：`(where name like ?, "%"+name)`
    - `sql_operator:"right like"`，表示：`(where name like ?, name+"%")`
    - `sql_operator:"full like"`，表示：`(where name like ?, "%"+name+"%")`
- 使用 `sql_expr` tag 来指示特殊表达式
    - `sql_expr:"$or"`，表示 `or`
        - `sql_field` tag 为空时生效
        - 若 `sql_field` tag 非空，则必须写成 `sql_field:"-"` 才生效
        - 类型必须是 `[]Where`，当前 Where 与 slice 元素之间使用 or 连接

#### 2.1.4 数据行筛选结构体 Update 特性

- 使用 `sql_field` tag 来指示映射列
    - `sql_field:"id"`， 表示 `where id = ?`
- 使用 `sql_expr` tag 指定表达式，目前支持：
    - `sql_expr:"+"`，表示：`update count = count + ?`
    - `sql_expr:"-"`，表示：`update count = count - ?`
    - `sql_expr:"merge_json"`，表示：
      ```sql
      update data = JSON_MERGE_PATCH(data, '{"key":"val"}')
      ```

### 2.2 定义业务 DAL

将 *GDAL 嵌入到业务 DAL 中，并指定三种业务结构体。

```go
package dal

import (
	"github.com/dirac-lee/gdal"
	"github.com/dirac-lee/gdal/example/dal/model"
	"gorm.io/gorm"
)

type UserDAL struct {
	*gdal.GDAL[model.User, model.UserWhere, model.UserUpdate]
}

func NewUserDAL(tx *gorm.DB) *UserDAL {
	return &UserDAL{
		gdal.NewGDAL[model.User, model.UserWhere, model.UserUpdate](tx),
	}
}
```

### 2.3 执行 CRUD

#### 2.3.1 初始化业务 DAL

```go
package main

import (
	"context"
	"github.com/dirac-lee/gdal/example/dal"
	"github.com/dirac-lee/gdal/example/dal/model"
	"github.com/dirac-lee/gdal/gutil/gptr"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

const (
	DemoDSN = "demo"
)

func main() {
	db, err := gorm.Open(mysql.Open(DemoDSN))
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	gdal := dal.NewUserDAL(db)

}
```

#### 2.3.2 新增

新增单条记录

```go
now := time.Now()
po := model.User{
    ID:         110,
    Name:       "dirac",
    Balance:    100,
    CreateTime: now,
    UpdateTime: now,
    Deleted:    false,
}
gdal.Create(ctx, &po)
```

新增多条记录

```go
now := time.Now()
pos := []*model.User{
    {
        ID:         120,
        Name:       "bob",
        Balance:    100,
        CreateTime: now,
        UpdateTime: now,
        Deleted:    false,
    },
    {
        ID:         130,
        Name:       "estele",
        Balance:    50,
        CreateTime: now,
        UpdateTime: now,
        Deleted:    false,
    },
}
gdal.MCreate(ctx, &pos)
```

#### 2.3.3 删除

物理删除

```go
where := &model.UserWhere{
    IDIn: []int64{110, 120},
}
gdal.Delete(ctx, where)
```

通过ID物理删除

```go
gdal.DeleteByID(ctx, 130)
```

#### 2.3.4 更新

更新

```go
where := &model.UserWhere{
    IDIn: []int64{110, 120},
}
update := &model.UserUpdate{
    BalanceAdd: gptr.Of[int64](10),
}
gdal.Update(ctx, where, update)
```

通过 ID 更新

```go
update := &model.UserUpdate{
    BalanceMinus: gptr.Of[int64](20),
}
gdal.UpdateByID(ctx, 130, update)
```

#### 2.3.5 查询

通用查询

```go
var pos []*model.User
where := &model.UserWhere{
    NameLike: gptr.Of("dirac"),
}
err = gdal.Find(ctx, &pos, where)
```

查询多条记录

```go
where := &model.UserWhere{
    IDIn: []int64{110, 120},
}
pos, err := gdal.MQuery(ctx, where)
```

分页查询1

```go
where := &model.UserWhere{
    IDIn: []int64{110, 120},
}
pos, total, err := gdal.MQueryByPaging(ctx, where, gptr.Of[int64](5), nil, gptr.Of("create_time desc"))
```

分页查询2

```go
where := &model.UserWhere{
    IDIn: []int64{110, 120},
}
pos, total, err := userDAL.MQueryByPagingOpt(ctx, where, gdal.WithLimit(5), gdal.WithOrder("create_time desc"))
```

#### 2.3.6 事务

```go
db.Transaction(func (tx *gorm.DB) error {
    
    update := &model.UserUpdate{
        BalanceMinus: gptr.Of[int64](20),
    }
    err := userDAL.WithTx(tx).UpdateByID(ctx, 130, update)
    if err != nil {
        return err // rollback
    }
    
    _, err = userDAL.WithTx(tx).DeleteByID(ctx, 130)
    if err != nil {
        return err // rollback
    }
    
    return nil // commit
})
```
# GDAL

[![go report card](https://goreportcard.com/badge/github.com/dirac-lee/gdal "go report card")](https://goreportcard.com/report/github.com/dirac-lee/gdal)
[![test status](https://github.com/dirac-lee/gdal/workflows/tests/badge.svg?branch=master "test status")](https://github.com/dirac-lee/gdal/actions)
[![MIT license](https://img.shields.io/badge/license-MIT-brightgreen.svg)](https://opensource.org/licenses/MIT)
[![Go.Dev reference](https://img.shields.io/badge/go.dev-reference-blue?logo=go&logoColor=white)](https://pkg.go.dev/github.com/dirac-lee/gdal?tab=doc)

A thoroughly object-relation mapping framework based on GORM. Struct object is all you need.

## Installation

> requirements: go >= 1.18

```bash
go get github.com/dirac-lee/gdal
```

## How To Use

Reference [example](./example) for details.

### 2.1 Definition of 3 business structs

### 2.1.1 Business structs

- PO：DB row persistent struct, mapped as a row of a DB table by tag `gorm`.
- Where：DB row selector struct, mapped as sql where condition by tag  `sql_field` and `sql_operator`.
- Update: DB row updater struct, mapped as sql update rule by tag  `sql_field` and `sql_expr`.

> ⚠️ Caution：the fields of Where 与 Update must be pointer. You can use [gptr](./gutil/gptr/gptr.go) to simplify the
> operation。

### 2.1.2 Examples of business structs

```go
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

#### 2.1.3 Features of Where

- Use tag `sql_field` tag to indicate the corresponding table column:
    - `sql_field:"id"`      ➡️  `where id = ?`
    - `sql_field:"dog.id"`  ➡️  `where dog.id = ?`
- Use tag `sql_operator` to indicate the selecting operator on the column:
    - `sql_operator:"="`   ➡️  `where name = ?`，默认
    - `sql_operator:">"`   ➡️  `where age > ? `
    - `sql_operator:">="`  ➡️  `where age >= ? `
    - `sql_operator:"<"`   ➡️  `where age < ? `
    - `sql_operator:"<="`  ➡️  `where age <= ? `
    - `sql_operator:"!="`  ➡️  `where age != ? `
    - `sql_operator:"null"`, the type of field must be `*bool`, and
        - `true`   ➡️  `where age is null`，
        - `false`  ➡️  `where age is not null`
    - `sql_operator:"in"`          ➡️  `where id in (?)`
      > ⚠️ Caution: empty slice `[]T{}` will be treated as nil slice `[]T(nil)`
    - `sql_operator:"not in"`      ➡️  `where id not in (?)`，
      > ⚠️ Caution: the same as `in`
    - `sql_operator:"like"`        ➡️  `(where name like ?, name)`
    - `sql_operator:"left like"`   ➡️  `(where name like ?, "%"+name)`
    - `sql_operator:"right like"`  ➡️  `(where name like ?, name+"%")`
    - `sql_operator:"full like"`   ➡️  `(where name like ?, "%"+name+"%")`
    - `sql_operator:"json_contains"` ➡️ "where JSON_CONTAINS(name, ?)" 
    - `sql_operator:"json_contains any"` ➡️ "where (JSON_CONTAINS(name, ?) or JSON_CONTAINS(name, ?))"
    - `sql_operator:"json_contains all"` ➡️ "where (JSON_CONTAINS(name, ?) and JSON_CONTAINS(name, ?))"
- Use tag `sql_expr` tag to indicate special expressions:
    - `sql_expr:"$or"`  ➡️  `or`
        - effective when there is no tag `sql_field`
        - if it has tag `sql_field`, the tag must be formed as `sql_field:"-"`
        - the type of field must be `[]Where`
        - connect current Where and elem of the `[]Where` with `or`

#### 2.1.4 Features of Update

- Use tag `sql_field` tag to indicate the corresponding table column:
    - `sql_field:"id"`  ➡️  `where id = ?`
- Use tag `sql_expr` to indicate special expressions:
    - `sql_expr:"+"`  ➡️  `update count = count + ?`
    - `sql_expr:"-"`  ➡️  `update count = count - ?`
    - `sql_expr:"json_set"`  ➡️ `update JSON_SET(data, $.attr, ?)`

### 2.2 Customize business DAL

Embed *GDAL into business DAL, and indicate the 3 business structs。

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

### 2.3 Execute CRUD

#### 2.3.1 Initialize business DAL

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

#### 2.3.2 Create

Create single record

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

Create multiple records

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

#### 2.3.3 Delete

Delete physically

```go
where := &model.UserWhere{
    IDIn: []int64{110, 120},
}
gdal.Delete(ctx, where)
```

Delete physically by ID

```go
gdal.DeleteByID(ctx, 130)
```

#### 2.3.4 Update

Update

```go
where := &model.UserWhere{
    IDIn: []int64{110, 120},
}
update := &model.UserUpdate{
    BalanceAdd: gptr.Of[int64](10),
}
gdal.Update(ctx, where, update)
```

Update by ID

```go
update := &model.UserUpdate{
    BalanceMinus: gptr.Of[int64](20),
}
gdal.UpdateByID(ctx, 130, update)
```

#### 2.3.5 Query

Query normally

```go
var pos []*model.User
where := &model.UserWhere{
    NameLike: gptr.Of("dirac"),
}
err = gdal.Find(ctx, &pos, where)
```

Query multiple records

```go
where := &model.UserWhere{
    IDIn: []int64{110, 120},
}
pos, err := gdal.MQuery(ctx, where)
```

Query multiple records by pagination (method 1)

```go
where := &model.UserWhere{
    IDIn: []int64{110, 120},
}
pos, total, err := gdal.MQueryByPaging(ctx, where, gptr.Of[int64](5), nil, gptr.Of("create_time desc"))
```

Query multiple records by pagination (method 2)

```go
where := &model.UserWhere{
    IDIn: []int64{110, 120},
}
pos, total, err := userDAL.MQueryByPagingOpt(ctx, where, gdal.WithLimit(5), gdal.WithOrder("create_time desc"))
```

#### 2.3.6 Transaction

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

### 2.3 Efficiency Features

#### 2.3.1 Inject Default

```go
// InjectDefault if you not set field `Deleted` of `UserWhere`,
// the where condition "deleted = false" will be automatically
// injected to the SQL.
func (where *UserWhere) InjectDefault() {
	where.Deleted = gptr.Of(false)
}
```

#### 2.3.2 Force Index

```go
// ForceIndex if field ID or IDIn is set, clause `USE INDEX "PRIMARY"`
// will be automatically injected to the SQL.
func (where UserWhere) ForceIndex() string {
    if where.ID != nil || len(where.IDIn) > 0 {
        return "PRIMARY" // USE INDEX "PRIMARY"
    }
    return "" // no USE INDEX
}
```

#### 2.3.3 Clauses

```go
now := time.Now()
hobbies, _ := json.Marshal([]string{"cooking", "coding"})
po := model.User{
ID:         110,
Name:       "dirac",
Balance:    100,
Hobbies:    string(hobbies),
CreateTime: now,
UpdateTime: now,
Deleted:    false,
}
err := userDAL.Clauses(clause.OnConflict{UpdateAll: true}).Create(ctx, &po)
fmt.Println(err)
```

will be mapped into SQL
```sql

```
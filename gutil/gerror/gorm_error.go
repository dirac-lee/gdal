package gerror

import (
	"errors"
	"gorm.io/gorm"
)

var (
	// ErrRecordNotFound record not found error
	ErrRecordNotFound = gorm.ErrRecordNotFound
	// ErrInvalidTransaction invalid transaction when you are trying to `Commit` or `Rollback`
	ErrInvalidTransaction = gorm.ErrInvalidTransaction
	// ErrNotImplemented not implemented
	ErrNotImplemented = gorm.ErrNotImplemented
	// ErrMissingWhereClause missing where clause
	ErrMissingWhereClause = gorm.ErrMissingWhereClause
	// ErrUnsupportedRelation unsupported relations
	ErrUnsupportedRelation = gorm.ErrUnsupportedRelation
	// ErrPrimaryKeyRequired primary keys required
	ErrPrimaryKeyRequired = gorm.ErrPrimaryKeyRequired
	// ErrModelValueRequired model value required
	ErrModelValueRequired = gorm.ErrModelValueRequired
	// ErrModelAccessibleFieldsRequired model accessible fields required
	ErrModelAccessibleFieldsRequired = gorm.ErrModelAccessibleFieldsRequired
	// ErrSubQueryRequired sub query required
	ErrSubQueryRequired = gorm.ErrSubQueryRequired
	// ErrInvalidData unsupported data
	ErrInvalidData = gorm.ErrInvalidData
	// ErrUnsupportedDriver unsupported driver
	ErrUnsupportedDriver = gorm.ErrUnsupportedDriver
	// ErrRegistered registered
	ErrRegistered = gorm.ErrRegistered
	// ErrInvalidField invalid field
	ErrInvalidField = gorm.ErrInvalidField
	// ErrEmptySlice empty slice found
	ErrEmptySlice = gorm.ErrEmptySlice
	// ErrDryRunModeUnsupported dry run mode unsupported
	ErrDryRunModeUnsupported = gorm.ErrDryRunModeUnsupported
	// ErrInvalidDB invalid db
	ErrInvalidDB = gorm.ErrInvalidDB
	// ErrInvalidValue invalid value
	ErrInvalidValue = gorm.ErrInvalidValue
	// ErrInvalidValueOfLength invalid values do not match length
	ErrInvalidValueOfLength = gorm.ErrInvalidValueOfLength
	// ErrPreloadNotAllowed preload is not allowed when count is used
	ErrPreloadNotAllowed = gorm.ErrPreloadNotAllowed
	// ErrDuplicatedKey occurs when there is a unique key constraint violation
	ErrDuplicatedKey = gorm.ErrDuplicatedKey
)

func IsRecordNotFoundError(err error) bool {
	return errors.Is(err, ErrRecordNotFound)
}

func IsErrInvalidTransaction(err error) bool {
	return errors.Is(err, ErrInvalidTransaction)
}

func IsErrNotImplemented(err error) bool {
	return errors.Is(err, ErrNotImplemented)
}

func IsErrMissingWhereClause(err error) bool {
	return errors.Is(err, ErrMissingWhereClause)
}

func IsErrUnsupportedRelation(err error) bool {
	return errors.Is(err, ErrUnsupportedRelation)
}

func IsErrPrimaryKeyRequired(err error) bool {
	return errors.Is(err, ErrPrimaryKeyRequired)
}

func IsErrModelValueRequired(err error) bool {
	return errors.Is(err, ErrModelValueRequired)
}

func IsErrModelAccessibleFieldsRequired(err error) bool {
	return errors.Is(err, ErrModelAccessibleFieldsRequired)
}

func IsErrSubQueryRequired(err error) bool {
	return errors.Is(err, ErrSubQueryRequired)
}

func IsErrInvalidData(err error) bool {
	return errors.Is(err, ErrInvalidData)
}

func IsErrUnsupportedDriver(err error) bool {
	return errors.Is(err, ErrUnsupportedDriver)
}

func IsErrRegistered(err error) bool {
	return errors.Is(err, ErrRegistered)
}

func IsErrInvalidField(err error) bool {
	return errors.Is(err, ErrInvalidField)
}

func IsErrEmptySlice(err error) bool {
	return errors.Is(err, ErrEmptySlice)
}

func IsErrDryRunModeUnsupported(err error) bool {
	return errors.Is(err, ErrDryRunModeUnsupported)
}

func IsErrInvalidDB(err error) bool {
	return errors.Is(err, ErrInvalidDB)
}

func IsErrInvalidValue(err error) bool {
	return errors.Is(err, ErrInvalidValue)
}

func IsErrInvalidValueOfLength(err error) bool {
	return errors.Is(err, ErrInvalidValueOfLength)
}

func IsErrPreloadNotAllowed(err error) bool {
	return errors.Is(err, ErrPreloadNotAllowed)
}

func IsErrDuplicatedKey(err error) bool {
	return errors.Is(err, ErrDuplicatedKey)
}

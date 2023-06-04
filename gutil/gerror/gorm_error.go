package gerror

import (
	"errors"
	"gorm.io/gorm"
)

func IsRecordNotFoundError(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}

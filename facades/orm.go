package facades

import (
	"github.com/nd-tools/capyvel/database"

	"gorm.io/gorm"
)

func Orm() *gorm.DB {
	return database.DB.Ctx
}

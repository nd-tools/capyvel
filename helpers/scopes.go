package helpers

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/nd-tools/capyvel/helpers/structaudit"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Error messages
var (
	ErrColumnNotValid = errors.New("the column not valid")
	ErrNameNotValid   = errors.New("field 'Name' not declared in the configurations")
)

func ScopeOrder(db *gorm.DB, fields []structaudit.FieldInfo, param string, desc bool) (*gorm.DB, error) {
	param = CleanText(param)
	if param != "" && len(fields) > 0 {
		var field *structaudit.FieldInfo
		for _, f := range fields {
			if param == f.TagJson || param == f.Name {
				field = &f
				break
			}
		}
		if field != nil {
			db = db.Order(clause.OrderByColumn{Column: clause.Column{Name: field.Name}, Desc: desc})
		} else {
			return db, ErrColumnNotValid
		}
	}
	return db, nil
}

func ScopeRelations(db *gorm.DB, relations []string, objType reflect.Type) (*gorm.DB, error) {
	if len(relations) > 0 {
		for _, relation := range relations {
			if _, err := structaudit.LocateFieldType(objType, relation, -1); err != nil {
				return nil, err
			}
			db = db.Preload(relation)
		}
	}
	return db, nil
}

func ScopePagination(page int, pageSize int, count int64) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		offset := (page) * pageSize
		return db.Offset(offset).Limit(pageSize)
	}
}

func ScopeSearch(db *gorm.DB, fields []structaudit.FieldInfo, param string) (*gorm.DB, error) {
	param = CleanText(param)
	if param != "" {
		var value = "%" + param + "%"
		var conditions []string
		var args []interface{}
		for _, f := range fields {
			if f.Name == "" {
				return db, ErrNameNotValid
			}
			conditions = append(conditions, fmt.Sprintf("%s LIKE ?", f.Name))
			args = append(args, value)
		}
		if len(conditions) > 0 {
			expr := strings.Join(conditions, " OR ")
			db = db.Where(fmt.Sprintf("(%s)", expr), args...)
		}
	}
	return db, nil
}

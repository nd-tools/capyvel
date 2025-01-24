package helpers

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/nd-tools/capyvel/database"
	"github.com/nd-tools/capyvel/helpers/structaudit"
	"github.com/nd-tools/capyvel/responses"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// NewOrm initializes a new Orm instance
func NewOrm() *Orm {
	return &Orm{
		db:   database.DB.Ctx,
		bind: *NewBind(),
	}
}

// Orm is the main struct for ORM operations
type Orm struct {
	db   *gorm.DB
	bind Bind
}

// FilterFunc defines a function type for filtering
type FilterFunc func(ctx *gin.Context, db *gorm.DB) (*gorm.DB, error)

// Configuration structs for various ORM operations
type (
	ListConfig struct {
		Db                *gorm.DB
		Limit             int
		DefaultOrderBy    string
		DefaultOrderDesc  bool
		ScanObj           bool
		DisablePagination bool
		SearchFields      []structaudit.FieldInfo
		OrderFields       []structaudit.FieldInfo
		FilterFunctions   []FilterFunc
	}
	AddConfig struct {
		Db          *gorm.DB
		BindMode    string
		WithAttach  bool
		DisableBind bool
		Batches     int
		BatchesSize int
	}
	UpdateConfig struct {
		Db                   *gorm.DB
		BindMode             string
		ColumnKey            string
		KeyParam             string
		BatchesSize          int
		WithAttach           bool
		DisableBind          bool
		DisableValidationKey bool
	}
	DeleteConfig struct {
		Db                   *gorm.DB
		ColumnKey            string
		KeyParam             string
		SoftDelete           bool
		DisableValidationKey bool
	}
	GetConfig struct {
		Db                   *gorm.DB
		ColumnKey            string
		KeyParam             string
		DisableRelations     bool
		DisableValidationKey bool
	}
	OrmParams struct {
		Search    string `form:"search,omitempty"`
		OrderBy   string `form:"orderBy,omitempty"`
		OrderDesc bool   `form:"orderDesc,omitempty"`
		Page      int    `form:"page,omitempty"`
		PageSize  int    `form:"pageSize,omitempty"`
	}
)

// Add creates a new record in the database
func (orm *Orm) Add(ctx *gin.Context, obj any, config AddConfig) (*responses.Api, *responses.Error) {
	if !config.DisableBind {
		if err := orm.bind.Json(ctx, ConfigJson{Obj: obj, Mode: config.BindMode}, nil); err != nil {
			return nil, &responses.Error{
				ErrorDetail: responses.ErrorDetail{
					Message: "error reading declared model",
					Error:   err,
					Type:    responses.TypeDB,
				},
				Code: http.StatusBadRequest,
			}
		}
	}
	db := config.Db
	if db == nil {
		db = orm.db
	}
	if !config.WithAttach {
		db = db.Omit(clause.Associations)
	} else {
		db = db.Session(&gorm.Session{FullSaveAssociations: true})
	}
	if config.BatchesSize > 0 {
		db.CreateBatchSize = config.BatchesSize
	} else {
		db.CreateBatchSize = -1
	}
	if structaudit.GetObjectKind(obj) == reflect.Slice {
		batches := 20
		if config.Batches > 0 {
			batches = config.Batches
		}
		if err := db.WithContext(ctx).CreateInBatches(obj, batches).Error; err != nil {
			return nil, &responses.Error{
				ErrorDetail: responses.ErrorDetail{
					Message: "error creating objects in the database",
					Error:   err,
					Type:    responses.TypeDB,
				},
				Code: http.StatusInternalServerError,
			}
		}
	} else {
		if err := db.WithContext(ctx).Create(obj).Error; err != nil {
			return nil, &responses.Error{
				ErrorDetail: responses.ErrorDetail{
					Message: "error creating object in the database",
					Error:   err,
					Type:    responses.TypeDB,
				},
				Code: http.StatusInternalServerError,
			}
		}
	}
	return &responses.Api{Data: obj}, nil
}

// Get retrieves a record from the database
func (orm *Orm) Get(ctx *gin.Context, obj any, config GetConfig) (*responses.Api, *responses.Error) {
	var param OrmParams
	if err := orm.bind.Url(ctx, ConfigUrl{Params: &param}); err != nil {
		return nil, &responses.Error{
			ErrorDetail: responses.ErrorDetail{
				Message: "error obtaining query params",
				Error:   err,
				Type:    responses.TypeDB,
			},
			Code: http.StatusBadRequest,
		}
	}

	db := config.Db
	if db == nil {
		db = orm.db
	}

	objType, err := structaudit.NormalizePointerType(obj)
	if err != nil {
		return nil, &responses.Error{
			ErrorDetail: responses.ErrorDetail{
				Message: "error normalizing received object",
				Error:   err,
				Type:    responses.TypeDB,
			},
			Code: http.StatusInternalServerError,
		}
	}

	var fieldInfo *structaudit.FieldInfo
	if config.ColumnKey != "" {
		f, err := structaudit.FindFieldInfoByName(objType, config.ColumnKey)
		if err != nil {
			return nil, &responses.Error{
				ErrorDetail: responses.ErrorDetail{
					Message: "error obtaining object information",
					Error:   err,
					Type:    responses.TypeDB,
				},
				Code: http.StatusInternalServerError,
			}
		}
		fieldInfo = f
	} else {
		f, err := structaudit.FindFieldInfoByTag(objType, "gorm", "primaryKey")
		if err != nil {
			return nil, &responses.Error{
				ErrorDetail: responses.ErrorDetail{
					Message: "error obtaining object information",
					Error:   err,
					Type:    responses.TypeDB,
				},
				Code: http.StatusInternalServerError,
			}
		}
		fieldInfo = f
	}
	if !config.DisableValidationKey {
		keyParam := "id"
		if config.KeyParam != "" {
			keyParam = config.KeyParam
		}
		if err := structaudit.ValidateFieldData(fieldInfo, ctx.Param(keyParam)); err != nil {
			return nil, &responses.Error{
				ErrorDetail: responses.ErrorDetail{
					Message: "error validating ID parameter",
					Error:   err,
					Type:    responses.TypeDB,
				},
				Code: http.StatusBadRequest,
			}
		}
	}
	if err := db.WithContext(ctx).First(obj, fieldInfo.Name+" = ?", fieldInfo.Value).Error; err != nil {
		return nil, &responses.Error{
			ErrorDetail: responses.ErrorDetail{
				Message: "error fetching object",
				Error:   err,
				Type:    responses.TypeDB,
			},
			Code: http.StatusInternalServerError,
		}
	}

	relations, _ := structaudit.ExtractFieldsByTag(objType, "gorm", "foreignKey")
	relationsMany, _ := structaudit.ExtractFieldsByTag(objType, "gorm", "many2many")
	relations = append(relations, relationsMany...)

	return &responses.Api{Data: obj, Relationships: relations}, nil
}

// Update modifies an existing record in the database
func (orm *Orm) Update(ctx *gin.Context, obj any, config UpdateConfig) (*responses.Api, *responses.Error) {
	db := config.Db
	if db == nil {
		db = orm.db
	}
	objType, err := structaudit.NormalizePointerType(obj)
	if err != nil {
		return nil, &responses.Error{
			ErrorDetail: responses.ErrorDetail{
				Message: "error normalizing received object",
				Error:   err,
				Type:    responses.TypeDB,
			},
			Code: http.StatusInternalServerError,
		}
	}
	fieldInfo, err := structaudit.FindFieldInfoByTag(objType, "gorm", "primaryKey")
	if err != nil {
		return nil, &responses.Error{
			ErrorDetail: responses.ErrorDetail{
				Message: "error finding primary key in the model",
				Error:   err,
				Type:    responses.TypeDB,
			},
			Code: http.StatusInternalServerError,
		}
	}
	if !config.DisableBind {
		if err := orm.bind.Json(ctx, ConfigJson{Obj: obj, Mode: config.BindMode}, nil); err != nil {
			return nil, &responses.Error{
				ErrorDetail: responses.ErrorDetail{
					Message: "error reading declared model",
					Error:   err,
					Type:    responses.TypeDB,
				},
				Code: http.StatusBadRequest,
			}
		}
	}
	if !config.DisableValidationKey {
		keyParam := "id"
		if config.KeyParam != "" {
			keyParam = config.KeyParam
		}
		if err := structaudit.ValidateFieldData(fieldInfo, ctx.Param(keyParam)); err != nil {
			return nil, &responses.Error{
				ErrorDetail: responses.ErrorDetail{
					Message: "error validating ID parameter",
					Error:   err,
					Type:    responses.TypeDB,
				},
				Code: http.StatusBadRequest,
			}
		}
	}
	if !config.WithAttach {
		db = db.Omit(clause.Associations)
	} else {
		db = db.Session(&gorm.Session{FullSaveAssociations: true})
	}
	if config.BatchesSize > 0 {
		db.CreateBatchSize = config.BatchesSize
	} else {
		db.CreateBatchSize = -1
	}
	if err := db.WithContext(ctx).Model(obj).Where(fieldInfo.Name+" = ?", fieldInfo.Value).Updates(obj).Error; err != nil {
		return nil, &responses.Error{
			ErrorDetail: responses.ErrorDetail{
				Message: "error updating object in the database",
				Error:   err,
				Type:    responses.TypeDB,
			},
			Code: http.StatusInternalServerError,
		}
	}
	return &responses.Api{Data: obj}, nil
}

// Delete removes a record from the database
func (orm *Orm) Delete(ctx *gin.Context, obj any, config DeleteConfig) (*responses.Api, *responses.Error) {
	db := config.Db
	if db == nil {
		db = orm.db
	}
	objType, err := structaudit.NormalizePointerType(obj)
	if err != nil {
		return nil, &responses.Error{
			ErrorDetail: responses.ErrorDetail{
				Message: "error normalizing received object",
				Error:   err,
				Type:    responses.TypeDB,
			},
			Code: http.StatusInternalServerError,
		}
	}
	fieldInfo, err := structaudit.FindFieldInfoByTag(objType, "gorm", "primaryKey")
	if err != nil {
		return nil, &responses.Error{
			ErrorDetail: responses.ErrorDetail{
				Message: "error finding primary key in the model",
				Error:   err,
				Type:    responses.TypeDB,
			},
			Code: http.StatusInternalServerError,
		}
	}
	if !config.DisableValidationKey {
		keyParam := "id"
		if config.KeyParam != "" {
			keyParam = config.KeyParam
		}
		if err := structaudit.ValidateFieldData(fieldInfo, ctx.Param(keyParam)); err != nil {
			return nil, &responses.Error{
				ErrorDetail: responses.ErrorDetail{
					Message: "error validating ID parameter",
					Error:   err,
					Type:    responses.TypeDB,
				},
				Code: http.StatusBadRequest,
			}
		}
	}
	if config.SoftDelete {
		if err := db.WithContext(ctx).Model(obj).Where(fieldInfo.Name+" = ?", fieldInfo.Value).Delete(obj).Error; err != nil {
			return nil, &responses.Error{
				ErrorDetail: responses.ErrorDetail{
					Message: "error soft deleting object",
					Error:   err,
					Type:    responses.TypeDB,
				},
				Code: http.StatusInternalServerError,
			}
		}
	} else {
		if err := db.WithContext(ctx).Unscoped().Where(fieldInfo.Name+" = ?", fieldInfo.Value).Delete(obj).Error; err != nil {
			return nil, &responses.Error{
				ErrorDetail: responses.ErrorDetail{
					Message: "error hard deleting object",
					Error:   err,
					Type:    responses.TypeDB,
				},
				Code: http.StatusInternalServerError,
			}
		}
	}
	return &responses.Api{Data: obj}, nil
}

// List retrieves multiple records from the database
func (orm *Orm) List(ctx *gin.Context, obj any, config ListConfig) (*responses.Api, *responses.Error) {
	var param OrmParams
	var err error
	if err := orm.bind.Url(ctx, ConfigUrl{Params: &param}); err != nil {
		return nil, &responses.Error{
			ErrorDetail: responses.ErrorDetail{
				Message: "error obtaining query params",
				Error:   err,
				Type:    responses.TypeDB,
			},
			Code: http.StatusBadRequest,
		}
	}

	db := config.Db
	if db == nil {
		db = orm.db
	}
	for _, filterFunction := range config.FilterFunctions {
		db, err = filterFunction(ctx, db)
		if err != nil {
			return nil, &responses.Error{ErrorDetail: responses.ErrorDetail{Message: "error en los 'Params query'", Error: err, Type: responses.TypeDB}, Code: http.StatusBadRequest}
		}
	}

	if config.SearchFields != nil {
		db, err = ScopeSearch(db, config.SearchFields, param.Search)
		if err != nil {
			return nil, &responses.Error{
				ErrorDetail: responses.ErrorDetail{
					Message: "error en los 'Params query'",
					Error:   err,
					Type:    responses.TypeDB,
				},
				Code: http.StatusBadRequest,
			}
		}
	}
	if !config.ScanObj {
		db = db.Model(obj)
	}
	totalRows := int64(0)
	if err := db.WithContext(ctx).Count(&totalRows).Error; err != nil {
		return nil, &responses.Error{
			ErrorDetail: responses.ErrorDetail{
				Message: "error counting total rows",
				Error:   err,
				Type:    responses.TypeDB,
			},
			Code: http.StatusInternalServerError,
		}
	}

	if config.DefaultOrderBy != "" {
		defaultOrderFieldRepeated := false
		for _, orderField := range config.OrderFields {
			if orderField.Name == config.DefaultOrderBy && param.OrderBy == config.DefaultOrderBy {
				defaultOrderFieldRepeated = true
			}
		}
		if !defaultOrderFieldRepeated {
			db = db.Order(clause.OrderByColumn{
				Column: clause.Column{Name: config.DefaultOrderBy},
				Desc:   config.DefaultOrderDesc,
			})
		}
	}
	if config.OrderFields != nil {
		db, err = ScopeOrder(db, config.OrderFields, param.OrderBy, param.OrderDesc)
		if err != nil {
			return nil, &responses.Error{
				ErrorDetail: responses.ErrorDetail{
					Message: "error en los 'Params query'",
					Error:   err,
					Type:    responses.TypeDB,
				},
				Code: http.StatusBadRequest,
			}
		}
	}

	if config.Limit == 0 {
		config.Limit = 30
	}

	if param.Page < 0 {
		param.Page = 0
	}
	if param.PageSize <= 0 {
		if config.Limit == -1 {
			param.PageSize = int(totalRows)
		} else if config.Limit > 0 {
			param.PageSize = config.Limit
		} else {
			param.PageSize = 30
		}
	} else {
		if param.PageSize > config.Limit {
			param.PageSize = config.Limit
		}
	}
	if !config.DisablePagination {
		db = db.WithContext(ctx).Scopes(ScopePagination(param.Page, param.PageSize, totalRows))
	}
	if config.ScanObj {
		if err := db.Scan(obj).Error; err != nil {
			return nil, &responses.Error{ErrorDetail: responses.ErrorDetail{Message: "error al escanear los registros", Error: err, Type: responses.TypeDB}, Code: http.StatusBadRequest}
		}
	} else {
		if err := db.Find(obj).Error; err != nil {
			return nil, &responses.Error{ErrorDetail: responses.ErrorDetail{Message: "error al escanear el modelo de los registros", Error: err, Type: responses.TypeDB}, Code: http.StatusBadRequest}
		}
	}
	baseURL := strings.TrimRight(ctx.Request.URL.Path, "/")
	base := fmt.Sprintf("%s?page=%d&pageSize=%d", baseURL, param.Page, param.PageSize)
	prev := fmt.Sprintf("%s?page=%d&pageSize=%d", baseURL, param.Page-1, param.PageSize)
	next := fmt.Sprintf("%s?page=%d&pageSize=%d", baseURL, param.Page+1, param.PageSize)
	meta := map[string]interface{}{
		"page":     param.Page,
		"pageSize": param.PageSize,
	}
	links := map[string]interface{}{
		"self": base,
		"next": next,
		"prev": prev,
	}
	return &responses.Api{Data: obj, Meta: meta, Links: links, TotalRows: totalRows}, nil
}

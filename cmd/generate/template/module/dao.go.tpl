package {{.DaoPackageName}}

import (
	{{- range .FieldImports}}
	"{{.}}"
	{{- end}}

	"github.com/morehao/golib/biz/genericdao"
	"gorm.io/gorm"
	"{{.ModulePath}}/{{.AppPathInProject}}/{{.ModelLayerName}}"
	"{{.ModulePath}}/pkg/dbclient"
)

type {{.StructName}}Cond struct {
	*genericdao.BaseCond
{{- range .ModelFields}}
{{- if not (isBuiltInField .FieldName)}}
	{{.FieldName}} {{.FieldType}}
{{- end}}
{{- end}}
}

func (c *{{.StructName}}Cond) BuildCondition(db *gorm.DB, tableName string) {
	if c.BaseCond != nil {
		c.BaseCond.BuildCondition(db, tableName)
	}
{{- range .ModelFields}}
{{- if not (isBuiltInField .FieldName)}}
{{- if isBasicType .FieldType}}
	{{- if eq .FieldType "string"}}
	if c.{{.FieldName}} != "" {
		db.Where("{{.ColumnName}} = ?", c.{{.FieldName}})
	}
	{{- else if eq .FieldType "time.Time"}}
	if !c.{{.FieldName}}.IsZero() {
		db.Where("{{.ColumnName}} = ?", c.{{.FieldName}})
	}
	{{- else}}
	if c.{{.FieldName}} != 0 {
		db.Where("{{.ColumnName}} = ?", c.{{.FieldName}})
	}
	{{- end}}
{{- else}}
{{- end}}
{{- end}}
{{- end}}
}

type {{.StructName}}Dao struct {
	*genericdao.GenericDao[{{.ModelLayerName}}.{{.StructName}}Entity, {{.ModelLayerName}}.{{.StructName}}EntityList]
}

func New{{.StructName}}Dao() *{{.StructName}}Dao {
	return &{{.StructName}}Dao{
		GenericDao: genericdao.NewGenericDao[{{.ModelLayerName}}.{{.StructName}}Entity, {{.ModelLayerName}}.{{.StructName}}EntityList](
			{{.ModelLayerName}}.TableName{{.StructName}}, "{{.StructName}}Dao",
			dbclient.{{.DBName}},
		),
	}
}
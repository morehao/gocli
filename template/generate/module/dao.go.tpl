package {{.DaoPackageName}}

import (
	{{- range .FieldImports}}
	"{{.}}"
	{{- end}}

	"github.com/morehao/golib/dbaccess/gormdao"
	"gorm.io/gorm"
	"{{.BaseModulePath}}/{{.AppModuleName}}/{{.ModelLayerName}}"
	"{{.BaseModulePath}}/pkg/dbclient"
)

type {{.StructName}}Cond struct {
	*gormdao.BaseCond[{{.PKFieldType}}]
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
		db.Where(tableName+".{{.ColumnName}} = ?", c.{{.FieldName}})
	}
	{{- else if eq .FieldType "time.Time"}}
	if !c.{{.FieldName}}.IsZero() {
		db.Where(tableName+".{{.ColumnName}} = ?", c.{{.FieldName}})
	}
	{{- else}}
	if c.{{.FieldName}} != 0 {
		db.Where(tableName+".{{.ColumnName}} = ?", c.{{.FieldName}})
	}
	{{- end}}
{{- else}}
{{- end}}
{{- end}}
{{- end}}
}

type {{.StructName}}Dao struct {
	*gormdao.Dao[{{.ModelLayerName}}.{{.StructName}}Entity, {{.ModelLayerName}}.{{.StructName}}EntityList, {{.PKFieldType}}]
}

func New{{.StructName}}Dao() *{{.StructName}}Dao {
	return &{{.StructName}}Dao{
		Dao: gormdao.NewDao[{{.ModelLayerName}}.{{.StructName}}Entity, {{.ModelLayerName}}.{{.StructName}}EntityList, {{.PKFieldType}}](
			{{.ModelLayerName}}.TableName{{.StructName}}, "{{.StructName}}Dao",
			dbclient.{{.DBName}},
		),
	}
}
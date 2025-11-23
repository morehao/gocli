package generate

import (
	"fmt"
	"os"
	"text/template"

	"github.com/morehao/golib/codegen"
	"github.com/morehao/golib/gutil"
)

const (
	nullableDefaultDesc = "not null"
	fieldDefaultKeyword = "default"
	fieldCommentKeyword = "comment"
)

func genModel() error {
	modelGenCfg := cfg.Model

	// 使用工具函数复制嵌入的模板文件到临时目录
	tplDir, getTplErr := CopyEmbeddedTemplatesToTempDir(TemplatesFS, "template/model")
	if getTplErr != nil {
		return getTplErr
	}
	// 清理临时目录
	defer os.RemoveAll(tplDir)

	analysisCfg := &codegen.ModuleCfg{
		CommonConfig: codegen.CommonConfig{
			PackageName:       modelGenCfg.PackageName,
			TplDir:            tplDir,
			RootDir:           workDir,
			LayerParentDirMap: cfg.LayerParentDirMap,
			LayerNameMap:      cfg.LayerNameMap,
			LayerPrefixMap:    cfg.LayerPrefixMap,
			TplFuncMap: template.FuncMap{
				TplFuncIsBuiltInField:      IsBuiltInField,
				TplFuncIsSysField:          IsSysField,
				TplFuncIsDefaultModelLayer: IsDefaultModelLayer,
				TplFuncIsDefaultDaoLayer:   IsDefaultDaoLayer,
			},
		},
		TableName: modelGenCfg.TableName,
	}
	gen := codegen.NewGenerator()
	analysisRes, analysisErr := gen.AnalysisModuleTpl(MysqlClient, analysisCfg)
	if analysisErr != nil {
		return fmt.Errorf("analysis model tpl error: %v", analysisErr)
	}

	// 如果配置了表名前缀，则从结构体名中去除前缀
	if modelGenCfg.TablePrefix != "" {
		analysisRes.StructName = RemoveTablePrefixFromStructName(
			analysisRes.StructName,
			analysisRes.TableName,
			modelGenCfg.TablePrefix,
		)
	}

	var modelLayerName, daoLayerName codegen.LayerName
	for _, v := range analysisRes.TplAnalysisList {
		if v.OriginLayerName == codegen.LayerNameModel {
			modelLayerName = v.LayerName
		}
		if v.OriginLayerName == codegen.LayerNameDao {
			daoLayerName = v.LayerName
		}
	}

	var genParamsList []codegen.GenParamsItem
	for _, v := range analysisRes.TplAnalysisList {
		var modelFields []ModelField
		for _, field := range v.ModelFields {
			nullableDesc := nullableDefaultDesc
			if field.IsNullable {
				nullableDesc = ""
			}
			defaultValue := fmt.Sprintf("%s %s", fieldDefaultKeyword, field.DefaultValue)
			if field.DefaultValue == "" {
				defaultValue = fmt.Sprintf("%s ''", fieldDefaultKeyword)
			}
			// GormComment 用于 model 层的 gorm tag，格式为 "comment: xxx"
			gormComment := fmt.Sprintf("%s: %s", fieldCommentKeyword, field.Comment)
			if field.Comment == "" {
				gormComment = ""
			}
			// Comment 用于 obj 层等其他地方的普通注释，直接使用原始注释
			comment := field.Comment
			modelFields = append(modelFields, ModelField{
				IsPrimaryKey:       field.ColumnKey == codegen.ColumnKeyPRI,
				FieldName:          gutil.ReplaceIdToID(field.FieldName),
				FieldLowerCaseName: gutil.SnakeToLowerCamel(field.FieldName),
				JsonTagName:        SnakeToLowerCamelWithID(field.ColumnName),
				FieldType:          field.FieldType,
				ColumnName:         field.ColumnName,
				ColumnType:         field.ColumnType,
				NullableDesc:       nullableDesc,
				DefaultValue:       defaultValue,
				GormComment:        gormComment,
				Comment:            comment,
			})
		}

		// 如果配置了表名前缀，则从文件名中去除前缀
		targetFilename := v.TargetFilename
		if modelGenCfg.TablePrefix != "" {
			targetFilename = RemoveTablePrefixFromFilename(
				v.TargetFilename,
				analysisRes.TableName,
				modelGenCfg.TablePrefix,
			)
		}

		genParamsList = append(genParamsList, codegen.GenParamsItem{
			TargetDir:      v.TargetDir,
			TargetFileName: targetFilename,
			Template:       v.Template,
			ExtraParams: ModelExtraParams{
				AppInfo: AppInfo{
					ProjectName:      cfg.appInfo.ProjectName,
					AppPathInProject: cfg.appInfo.AppPathInProject,
					AppName:          cfg.appInfo.AppName,
					ProjectRootPath:  cfg.appInfo.ProjectRootPath,
					ModulePath:       cfg.appInfo.ModulePath,
				},
				PackageName:    analysisRes.PackageName,
				TableName:      analysisRes.TableName,
				ModelLayerName: string(modelLayerName),
				DaoLayerName:   string(daoLayerName),
				Description:    modelGenCfg.Description,
				StructName:     analysisRes.StructName,
				Template:       v.Template,
				ModelFields:    modelFields,
			},
		})

	}
	genParams := &codegen.GenParams{
		ParamsList: genParamsList,
	}
	if err := gen.Gen(genParams); err != nil {
		return err
	}
	return nil
}

type ModelField struct {
	IsPrimaryKey       bool   // 是否是主键
	FieldName          string // 字段名称
	FieldLowerCaseName string // 字段名称小驼峰
	JsonTagName        string // JSON 标签名称，特殊处理 _id 后缀为 ID
	FieldType          string // 字段数据类型，如int、string
	ColumnName         string // 列名
	ColumnType         string // 列数据类型，如varchar(255)
	NullableDesc       string // 是否允许为空描述，如 NOT NULL
	DefaultValue       string // 默认值,如 DEFAULT 0
	GormComment        string // gorm tag中的注释，格式为 "comment: xxx"，用于 model 层
	Comment            string // 普通注释，用于 obj 层等其他地方
}

type ModelExtraParams struct {
	AppInfo
	PackageName    string
	ModelLayerName string
	DaoLayerName   string
	TableName      string
	Description    string
	StructName     string
	Template       *template.Template
	ModelFields    []ModelField
}

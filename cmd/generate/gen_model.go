package generate

import (
	"fmt"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"text/template"

	"github.com/morehao/golib/codegen"
	"github.com/morehao/golib/gast"
	"github.com/morehao/golib/gutil"
)

const (
	nullableDefaultDesc = "not null"
	fieldDefaultKeyword = "default"
	fieldCommentKeyword = "comment"
)

func genModel() error {
	modelGenCfg := cfg.Model

	fmt.Printf("[Model] Generating model based on table: %s\n", modelGenCfg.TableName)

	// 使用工具函数复制嵌入的模板文件到临时目录
	tplDir, getTplErr := CopyEmbeddedTemplatesToTempDir(TemplatesFS, "template/model")
	if getTplErr != nil {
		return getTplErr
	}
	// 清理临时目录
	defer os.RemoveAll(tplDir)

	layerNameMap := buildLayerNameMap(cfg.ServiceName)

	analysisCfg := &codegen.ModuleCfg{
		CommonConfig: codegen.CommonConfig{
			PackageName:       modelGenCfg.PackageName,
			TplDir:            tplDir,
			RootDir:           workDir,
			LayerParentDirMap: defaultLayerParentDirMap,
			LayerNameMap:      layerNameMap,
			LayerPrefixMap:    defaultLayerPrefixMap,
			TplFuncMap: template.FuncMap{
				TplFuncIsBuiltInField:      IsBuiltInField,
				TplFuncIsSysField:          IsSysField,
				TplFuncIsDefaultModelLayer: IsDefaultModelLayer,
				TplFuncIsDefaultDaoLayer:   IsDefaultDaoLayer,
				TplFuncHasTimeField:        HasTimeField,
				TplFuncGetFieldImports:     GetFieldImports,
				TplFuncIsBasicType:         IsBasicType,
			},
		},
		TableName: modelGenCfg.TableName,
	}

	gen := codegen.NewGenerator()
	analysisRes, analysisErr := gen.AnalysisModuleTpl(DBClient, analysisCfg)
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
	var tableLayerItem *codegen.ModuleTplAnalysisItem
	var modelTargetDir string
	for _, v := range analysisRes.TplAnalysisList {
		if v.OriginLayerName == codegen.LayerName("table") {
			tmpV := v
			tableLayerItem = &tmpV
			continue
		}

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
				IsPrimaryKey:         field.ColumnKey == codegen.ColumnKeyPRI,
				FieldName:            gutil.ReplaceIdToID(field.FieldName),
				FieldLowerCaseName:   gutil.SnakeToLowerCamel(field.FieldName),
				JsonTagName:          SnakeToLowerCamelWithID(field.ColumnName),
				FieldType:            field.FieldType,
				ColumnName:           field.ColumnName,
				ColumnType:           field.ColumnType,
				NullableDesc:         nullableDesc,
				DefaultValue:         defaultValue,
				GormComment:          gormComment,
				Comment:              comment,
				StructNameLowerCamel: gutil.FirstLetterToLower(analysisRes.StructName),
				IndexName:            field.IndexName,
				IsUniqueIndex:        field.IsUniqueIndex,
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

		targetDir := v.TargetDir
		if v.OriginLayerName == codegen.LayerNameDao {
			targetDir = filepath.Dir(v.TargetDir)
		}
		if v.OriginLayerName == codegen.LayerNameModel {
			modelTargetDir = targetDir
		}

		fieldImports := calcFieldImports(modelFields)
		if v.OriginLayerName == codegen.LayerNameObject {
			fieldImports = calcFieldImports(modelFields, "time")
		}
		genParamsList = append(genParamsList, codegen.GenParamsItem{
			TargetDir:      targetDir,
			TargetFileName: targetFilename,
			Template:       v.Template,
			ExtraParams: ModelExtraParams{
				AppInfo: AppInfo{
					ProjectName:     cfg.appInfo.ProjectName,
					AppName:         cfg.appInfo.AppName,
					ProjectRootPath: cfg.appInfo.ProjectRootPath,
					BaseModulePath:  cfg.appInfo.BaseModulePath,
					AppModuleName:   cfg.appInfo.AppModuleName,
				},
				PackageName:    analysisRes.PackageName,
				TableName:      analysisRes.TableName,
				ModelLayerName: string(modelLayerName),
				DaoLayerName:   string(daoLayerName),
				DaoPackageName: string(daoLayerName),
				DBName:         fmt.Sprintf("%sDB", gutil.FirstLetterToUpper(cfg.ServiceName)),
				Description:    modelGenCfg.Description,
				StructName:     analysisRes.StructName,
				Template:       v.Template,
				ModelFields:    modelFields,
				FieldImports:   fieldImports,
			},
		})

	}
	genParams := &codegen.GenParams{
		ParamsList: genParamsList,
	}
	if err := gen.Gen(genParams); err != nil {
		return err
	}

	if tableLayerItem != nil {
		constName := fmt.Sprintf("TableName%s", analysisRes.StructName)
		tableFilepath := filepath.Join(modelTargetDir, "table.go")
		if gutil.FileExists(tableFilepath) {
			if err := gast.AddConstToFile(tableFilepath, constName, analysisRes.TableName, token.STRING); err != nil {
				return fmt.Errorf("failed to append table const: %v", err)
			}
		} else {
			tableExtraParams := ModelExtraParams{
				AppInfo: AppInfo{
					ProjectName:     cfg.appInfo.ProjectName,
					AppName:         cfg.appInfo.AppName,
					ProjectRootPath: cfg.appInfo.ProjectRootPath,
					BaseModulePath:  cfg.appInfo.BaseModulePath,
					AppModuleName:   cfg.appInfo.AppModuleName,
				},
				PackageName:    analysisRes.PackageName,
				TableName:      analysisRes.TableName,
				ModelLayerName: string(modelLayerName),
				StructName:     analysisRes.StructName,
			}
			tableGenParams := &codegen.GenParams{
				ParamsList: []codegen.GenParamsItem{
					{
						TargetDir:      modelTargetDir,
						TargetFileName: "table.go",
						Template:       tableLayerItem.Template,
						ExtraParams:    tableExtraParams,
					},
				},
			}
			if err := gen.Gen(tableGenParams); err != nil {
				return fmt.Errorf("failed to generate table.go: %v", err)
			}
		}
	}

	fmt.Printf("[Model] Generated layers: model(%s), dao(%s)\n", modelLayerName, daoLayerName)
	return nil
}

type ModelField struct {
	IsPrimaryKey         bool   // 是否是主键
	FieldName            string // 字段名称
	FieldLowerCaseName   string // 字段名称小驼峰
	JsonTagName          string // JSON 标签名称，特殊处理 _id 后缀为 ID
	FieldType            string // 字段数据类型，如int、string
	ColumnName           string // 列名
	ColumnType           string // 列数据类型，如varchar(255)
	NullableDesc         string // 是否允许为空描述，如 NOT NULL
	DefaultValue         string // 默认值,如 DEFAULT 0
	GormComment          string // gorm tag中的注释，格式为 "comment: xxx"，用于 model 层
	Comment              string // 普通注释，用于 obj 层等其他地方
	StructNameLowerCamel string // 结构体名称小驼峰，用于模板引用
	IndexName            string // 索引名称
	IsUniqueIndex        bool   // 是否唯一索引
}

type ModelExtraParams struct {
	AppInfo
	PackageName    string
	ModelLayerName string
	DaoLayerName   string
	DaoPackageName string
	DBName         string
	TableName      string
	Description    string
	StructName     string
	Template       *template.Template
	ModelFields    []ModelField
	FieldImports   []string
}

func calcFieldImports(fields []ModelField, excludeImports ...string) []string {
	importMap := make(map[string]struct{})
	for _, field := range fields {
		if IsBuiltInField(field.FieldName) {
			continue
		}
		if importInfo, ok := fieldTypeImportMap[field.FieldType]; ok {
			importMap[importInfo.ImportPath] = struct{}{}
		}
	}
	for _, exclude := range excludeImports {
		delete(importMap, exclude)
	}
	imports := make([]string, 0, len(importMap))
	for path := range importMap {
		imports = append(imports, path)
	}
	sort.Strings(imports)
	return imports
}

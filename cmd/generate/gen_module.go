package generate

import (
	"fmt"
	"go/token"
	"os"
	"path/filepath"
	"text/template"

	"github.com/morehao/golib/codegen"
	"github.com/morehao/golib/gast"
	"github.com/morehao/golib/gutil"
)

func genModule() error {
	moduleGenCfg := cfg.Module

	fmt.Printf("[Module] Generating module based on table: %s\n", moduleGenCfg.TableName)

	// 使用工具函数复制嵌入的模板文件到临时目录
	tplDir, getTplErr := CopyEmbeddedTemplatesToTempDir(TemplatesFS, "template/module")
	if getTplErr != nil {
		return getTplErr
	}
	// 清理临时目录
	defer os.RemoveAll(tplDir)

	layerNameMap := buildLayerNameMap(cfg.ServiceName)

	analysisCfg := &codegen.ModuleCfg{
		CommonConfig: codegen.CommonConfig{
			PackageName:       moduleGenCfg.PackageName,
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
		TableName: moduleGenCfg.TableName,
	}
	gen := codegen.NewGenerator()
	analysisRes, analysisErr := gen.AnalysisModuleTpl(DBClient, analysisCfg)
	if analysisErr != nil {
		return fmt.Errorf("analysis module tpl error: %v", analysisErr)
	}

	// 如果配置了表名前缀，则从结构体名中去除前缀
	if moduleGenCfg.TablePrefix != "" {
		analysisRes.StructName = RemoveTablePrefixFromStructName(
			analysisRes.StructName,
			analysisRes.TableName,
			moduleGenCfg.TablePrefix,
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
	appInfo := cfg.appInfo

	var genParamsList []codegen.GenParamsItem
	var codeLayerItem *codegen.ModuleTplAnalysisItem
	var tableLayerItem *codegen.ModuleTplAnalysisItem
	var modelTargetDir string
	for _, v := range analysisRes.TplAnalysisList {
		if v.OriginLayerName == codegen.LayerNameCode {
			tmpV := v
			codeLayerItem = &tmpV
			continue
		}
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
		if moduleGenCfg.TablePrefix != "" {
			targetFilename = RemoveTablePrefixFromFilename(
				v.TargetFilename,
				analysisRes.TableName,
				moduleGenCfg.TablePrefix,
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
				ExtraParams: ModuleExtraParams{
					AppInfo: AppInfo{
						ProjectName:     appInfo.ProjectName,
						AppName:         appInfo.AppName,
						ProjectRootPath: appInfo.ProjectRootPath,
						BaseModulePath:  appInfo.BaseModulePath,
						AppModuleName:   appInfo.AppModuleName,
					},
					PackageName:          analysisRes.PackageName,
					TableName:            analysisRes.TableName,
					ModelLayerName:       string(modelLayerName),
					DaoLayerName:         string(daoLayerName),
					DaoPackageName:       string(daoLayerName),
					DBName:               fmt.Sprintf("%sDB", gutil.FirstLetterToUpper(cfg.ServiceName)),
					Description:          moduleGenCfg.Description,
					StructName:           analysisRes.StructName,
					StructNameLowerCamel: gutil.FirstLetterToLower(analysisRes.StructName),
					Template:             v.Template,
					ModelFields:          modelFields,
					FieldImports:         fieldImports,
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
			tableExtraParams := ModuleExtraParams{
				AppInfo: AppInfo{
					ProjectName:     appInfo.ProjectName,
					AppName:         appInfo.AppName,
					ProjectRootPath: appInfo.ProjectRootPath,
					BaseModulePath:  appInfo.BaseModulePath,
					AppModuleName:   appInfo.AppModuleName,
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

	// 注册路由
	routerContent := fmt.Sprintf("%sRouter(groups)", gutil.FirstLetterToLower(analysisRes.StructName))
	routerEnterFilepath := filepath.Join(workDir, "/internal/router/router.go")
	if err := gast.AddContentToFunc(routerEnterFilepath, "RegisterRouter", routerContent); err != nil {
		return fmt.Errorf("router appendContentToFunc error: %v", err)
	}
	fmt.Printf("[Module] Registered router: %sRouter\n", gutil.FirstLetterToLower(analysisRes.StructName))

	// 处理code层：生成错误码文件到项目根目录的pkg/code目录
	if codeLayerItem != nil {
		// 构造code层的ExtraParams
		var modelFields []ModelField
		for _, field := range codeLayerItem.ModelFields {
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

		codeExtraParams := ModuleExtraParams{
			AppInfo: AppInfo{
				ProjectName:     appInfo.ProjectName,
				AppName:         appInfo.AppName,
				ProjectRootPath: appInfo.ProjectRootPath,
				BaseModulePath:  appInfo.BaseModulePath,
				AppModuleName:   appInfo.AppModuleName,
			},
			PackageName:          analysisRes.PackageName,
			TableName:            analysisRes.TableName,
			ModelLayerName:       string(modelLayerName),
			DaoLayerName:         string(daoLayerName),
			DaoPackageName:       string(daoLayerName),
			DBName:               fmt.Sprintf("%sDB", gutil.FirstLetterToUpper(appInfo.AppName)),
			Description:          moduleGenCfg.Description,
			StructName:           analysisRes.StructName,
			StructNameLowerCamel: gutil.FirstLetterToLower(analysisRes.StructName),
			Template:             codeLayerItem.Template,
			ModelFields:          modelFields,
			FieldImports:         calcFieldImports(modelFields),
		}

		// 生成错误码文件到项目根目录的pkg/code目录
		codeTargetDir := filepath.Join(cfg.appInfo.ProjectRootPath, "pkg/code")
		codeTargetFileName := fmt.Sprintf("%s.go", moduleGenCfg.PackageName)

		// 确保目录存在
		if err := os.MkdirAll(codeTargetDir, 0755); err != nil {
			return fmt.Errorf("failed to create code directory: %v", err)
		}

		// 使用codegen的createFile函数生成文件（支持追加）
		codeGenParams := &codegen.GenParams{
			ParamsList: []codegen.GenParamsItem{
				{
					TargetDir:      codeTargetDir,
					TargetFileName: codeTargetFileName,
					Template:       codeLayerItem.Template,
					ExtraParams:    codeExtraParams,
				},
			},
		}
		if err := gen.Gen(codeGenParams); err != nil {
			return fmt.Errorf("failed to generate code file: %v", err)
		}

		// 注册错误码到项目根目录的pkg/code/code.go
		codeContent := fmt.Sprintf("registerError(%sErrorMsgMap)", gutil.FirstLetterToLower(analysisRes.StructName))
		codeEnterFilepath := filepath.Join(cfg.appInfo.ProjectRootPath, "pkg/code/code.go")
		if err := gast.AddContentToFunc(codeEnterFilepath, "init", codeContent); err != nil {
			return fmt.Errorf("code appendContentToFunc error: %v", err)
		}
		fmt.Printf("[Module] Registered error code: %sErrorMsgMap\n", gutil.FirstLetterToLower(analysisRes.StructName))
	}

	fmt.Printf("[Module] Generated layers: model(%s), dao(%s)\n", modelLayerName, daoLayerName)
	return nil
}

type ModuleExtraParams struct {
	AppInfo
	PackageName          string
	ModelLayerName       string
	DaoLayerName         string
	DaoPackageName       string
	DBName               string
	TableName            string
	Description          string
	StructName           string
	StructNameLowerCamel string
	Template             *template.Template
	ModelFields          []ModelField
	FieldImports         []string
}

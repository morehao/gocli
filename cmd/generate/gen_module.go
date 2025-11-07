package generate

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/morehao/golib/codegen"
	"github.com/morehao/golib/gast"
	"github.com/morehao/golib/gutils"
)

func genModule() error {
	moduleGenCfg := cfg.Module

	// 使用工具函数复制嵌入的模板文件到临时目录
	tplDir, getTplErr := CopyEmbeddedTemplatesToTempDir(TemplatesFS, "template/module")
	if getTplErr != nil {
		return getTplErr
	}
	// 清理临时目录
	defer os.RemoveAll(tplDir)

	analysisCfg := &codegen.ModuleCfg{
		CommonConfig: codegen.CommonConfig{
			PackageName:       moduleGenCfg.PackageName,
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
		TableName: moduleGenCfg.TableName,
	}
	gen := codegen.NewGenerator()
	analysisRes, analysisErr := gen.AnalysisModuleTpl(MysqlClient, analysisCfg)
	if analysisErr != nil {
		return fmt.Errorf("analysis module tpl error: %v", analysisErr)
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
	var codeLayerItem *codegen.ModuleTplAnalysisItem
	for _, v := range analysisRes.TplAnalysisList {
		// 如果是code层，单独处理，不加入通用生成列表
		if v.OriginLayerName == codegen.LayerNameCode {
			tmpV := v
			codeLayerItem = &tmpV
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
			comment := fmt.Sprintf("%s: %s", fieldCommentKeyword, field.Comment)
			if field.Comment == "" {
				comment = ""
			}
			modelFields = append(modelFields, ModelField{
				IsPrimaryKey:       field.ColumnKey == codegen.ColumnKeyPRI,
				FieldName:          gutils.ReplaceIdToID(field.FieldName),
				FieldLowerCaseName: gutils.SnakeToLowerCamel(field.FieldName),
				FieldType:          field.FieldType,
				ColumnName:         field.ColumnName,
				ColumnType:         field.ColumnType,
				NullableDesc:       nullableDesc,
				DefaultValue:       defaultValue,
				Comment:            comment,
			})
		}

		genParamsList = append(genParamsList, codegen.GenParamsItem{
			TargetDir:      v.TargetDir,
			TargetFileName: v.TargetFilename,
			Template:       v.Template,
			ExtraParams: ModuleExtraParams{
				AppInfo: AppInfo{
					ProjectName:      cfg.appInfo.ProjectName,
					AppPathInProject: cfg.appInfo.AppPathInProject,
					AppName:          cfg.appInfo.AppName,
				},
				PackageName:          analysisRes.PackageName,
				TableName:            analysisRes.TableName,
				ModelLayerName:       string(modelLayerName),
				DaoLayerName:         string(daoLayerName),
				Description:          moduleGenCfg.Description,
				StructName:           analysisRes.StructName,
				StructNameLowerCamel: gutils.FirstLetterToLower(analysisRes.StructName),
				Template:             v.Template,
				ModelFields:          modelFields,
			},
		})

	}
	genParams := &codegen.GenParams{
		ParamsList: genParamsList,
	}
	if err := gen.Gen(genParams); err != nil {
		return err
	}

	// 注册路由
	routerContent := fmt.Sprintf("%sRouter(routerGroup)", gutils.FirstLetterToLower(analysisRes.StructName))
	routerEnterFilepath := filepath.Join(workDir, "/router/enter.go")
	if err := gast.AddContentToFunc(routerEnterFilepath, "RegisterRouter", routerContent); err != nil {
		return fmt.Errorf("router appendContentToFunc error: %v", err)
	}

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
			comment := fmt.Sprintf("%s: %s", fieldCommentKeyword, field.Comment)
			if field.Comment == "" {
				comment = ""
			}
			modelFields = append(modelFields, ModelField{
				IsPrimaryKey:       field.ColumnKey == codegen.ColumnKeyPRI,
				FieldName:          gutils.ReplaceIdToID(field.FieldName),
				FieldLowerCaseName: gutils.SnakeToLowerCamel(field.FieldName),
				FieldType:          field.FieldType,
				ColumnName:         field.ColumnName,
				ColumnType:         field.ColumnType,
				NullableDesc:       nullableDesc,
				DefaultValue:       defaultValue,
				Comment:            comment,
			})
		}

		codeExtraParams := ModuleExtraParams{
			AppInfo: AppInfo{
				ProjectName:      cfg.appInfo.ProjectName,
				AppPathInProject: cfg.appInfo.AppPathInProject,
				AppName:          cfg.appInfo.AppName,
			},
			PackageName:          analysisRes.PackageName,
			TableName:            analysisRes.TableName,
			ModelLayerName:       string(modelLayerName),
			DaoLayerName:         string(daoLayerName),
			Description:          moduleGenCfg.Description,
			StructName:           analysisRes.StructName,
			StructNameLowerCamel: gutils.FirstLetterToLower(analysisRes.StructName),
			Template:             codeLayerItem.Template,
			ModelFields:          modelFields,
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

		// 注册错误码到项目根目录的pkg/code/enter.go
		codeContent := fmt.Sprintf("registerError(%sErrorMsgMap)", gutils.FirstLetterToLower(analysisRes.StructName))
		codeEnterFilepath := filepath.Join(cfg.appInfo.ProjectRootPath, "pkg/code/enter.go")
		if err := gast.AddContentToFunc(codeEnterFilepath, "init", codeContent); err != nil {
			return fmt.Errorf("code appendContentToFunc error: %v", err)
		}
	}

	return nil
}

type ModuleExtraParams struct {
	AppInfo
	PackageName          string
	ModelLayerName       string
	DaoLayerName         string
	TableName            string
	Description          string
	StructName           string
	StructNameLowerCamel string // 结构体小写驼峰名
	Template             *template.Template
	ModelFields          []ModelField
}

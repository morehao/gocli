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

func genApi() error {
	apiGenCfg := cfg.Api

	// 使用工具函数复制嵌入的模板文件到临时目录
	tplDir, getTplErr := CopyEmbeddedTemplatesToTempDir(TemplatesFS, "template/api")
	if getTplErr != nil {
		return getTplErr
	}
	// 清理临时目录
	defer os.RemoveAll(tplDir)

	analysisCfg := &codegen.ApiCfg{
		CommonConfig: codegen.CommonConfig{
			PackageName:       apiGenCfg.PackageName,
			TplDir:            tplDir,
			RootDir:           workDir,
			LayerParentDirMap: cfg.LayerParentDirMap,
			LayerNameMap:      cfg.LayerNameMap,
			LayerPrefixMap:    cfg.LayerPrefixMap,
		},
		TargetFilename: apiGenCfg.TargetFilename,
	}
	gen := codegen.NewGenerator()
	analysisRes, analysisErr := gen.AnalysisApiTpl(analysisCfg)
	if analysisErr != nil {
		return fmt.Errorf("analysis api tpl error: %v", analysisErr)
	}

	structName := gutils.SnakeToPascal(gutils.TrimFileExtension(apiGenCfg.TargetFilename))
	structNameLowerCamel := gutils.FirstLetterToLower(structName)
	functionName := gutils.FirstLetterToUpper(apiGenCfg.FunctionName)
	functionNameLowerCamel := gutils.FirstLetterToLower(apiGenCfg.FunctionName)
	var genParamsList []codegen.GenParamsItem
	var isNewRouter, isNewController bool
	var controllerFilepath, serviceFilepath string
	for _, v := range analysisRes.TplAnalysisList {
		switch v.LayerName {
		case codegen.LayerNameRouter:
			if v.TargetFileExist {
				goFilepath := filepath.Join(v.TargetDir, v.TargetFilename)
				funcName := fmt.Sprintf("%sRouter", structNameLowerCamel)
				_, hasFunc, findFuncErr := gast.FindFunction(goFilepath, funcName)
				if findFuncErr != nil {
					return fmt.Errorf("find function error: %v", findFuncErr)
				}
				isNewRouter = !hasFunc
			} else {
				isNewRouter = true
			}
		case codegen.LayerNameController:
			controllerFilepath = filepath.Join(v.TargetDir, v.TargetFilename)
			isNewController = !v.TargetFileExist
		case codegen.LayerNameService:
			serviceFilepath = filepath.Join(v.TargetDir, v.TargetFilename)
		}

		genParamsList = append(genParamsList, codegen.GenParamsItem{
			TargetDir:      v.TargetDir,
			TargetFileName: v.TargetFilename,
			Template:       v.Template,
			ExtraParams: ApiExtraParams{
				AppInfo: AppInfo{
					ProjectName:      cfg.appInfo.ProjectName,
					AppPathInProject: cfg.appInfo.AppPathInProject,
					AppName:          cfg.appInfo.AppName,
					ProjectRootPath:  cfg.appInfo.ProjectRootPath,
					ModulePath:       cfg.appInfo.ModulePath,
				},
				PackageName:            analysisRes.PackageName,
				TargetFileExist:        v.TargetFileExist,
				IsNewRouter:            isNewRouter,
				Description:            apiGenCfg.Description,
				StructName:             structName,
				StructNameLowerCamel:   structNameLowerCamel,
				FunctionName:           functionName,
				FunctionNameLowerCamel: functionNameLowerCamel,
				HttpMethod:             apiGenCfg.HttpMethod,
				ApiDocTag:              apiGenCfg.ApiDocTag,
				Template:               v.Template,
			},
		})

	}
	genParams := &codegen.GenParams{
		ParamsList: genParamsList,
	}
	if err := gen.Gen(genParams); err != nil {
		return err
	}

	if !isNewController {
		// 将方法添加到interface接口中
		controllerInterfaceName := fmt.Sprintf("%sCtr", structName)
		if err := gast.AddMethodToInterface(controllerFilepath, structNameLowerCamel+"Ctr", functionName, controllerInterfaceName); err != nil {
			return fmt.Errorf("add controller method to interface error: %w", err)
		}
		serviceInterfaceName := fmt.Sprintf("%sSvc", structName)
		if err := gast.AddMethodToInterface(serviceFilepath, structNameLowerCamel+"Svc", functionName, serviceInterfaceName); err != nil {
			return fmt.Errorf("add service method to interface error: %w", err)
		}
	}

	// 	注册路由
	if isNewRouter {
		routerCallContent := fmt.Sprintf("%sRouter(routerGroup)", structNameLowerCamel)
		routerEnterFilepath := filepath.Join(workDir, "/router/enter.go")
		if err := gast.AddContentToFunc(routerEnterFilepath, "RegisterRouter", routerCallContent); err != nil {
			return fmt.Errorf("new router appendContentToFunc error: %v", err)
		}
	} else {
		routerCallContent := fmt.Sprintf(`routerGroup.%s("/%s", %sCtr.%s) // %s`, apiGenCfg.HttpMethod, functionNameLowerCamel, structNameLowerCamel, functionName, apiGenCfg.Description)
		routerEnterFilepath := filepath.Join(workDir, fmt.Sprintf("/router/%s.go", gutils.TrimFileExtension(apiGenCfg.PackageName)))
		if err := gast.AddContentToFuncWithLineNumber(routerEnterFilepath, fmt.Sprintf("%sRouter", structNameLowerCamel), routerCallContent, -2); err != nil {
			return fmt.Errorf("appendContentToFunc error: %v", err)
		}
	}
	return nil
}

type ApiExtraParams struct {
	AppInfo
	PackageName            string
	Description            string
	TargetFileExist        bool
	IsNewRouter            bool
	HttpMethod             string
	StructName             string
	StructNameLowerCamel   string
	FunctionName           string
	FunctionNameLowerCamel string
	ApiDocTag              string
	Template               *template.Template
}

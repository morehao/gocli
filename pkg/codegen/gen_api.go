package codegen

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/morehao/golib/codegen"
	"github.com/morehao/golib/gast"
	"github.com/morehao/golib/gutil"
)

func (g *Generator) genApi() error {
	apiGenCfg := g.cfg.Api

	// 使用工具函数复制嵌入的模板文件到临时目录
	tplDir, getTplErr := CopyEmbeddedTemplatesToTempDir(TemplatesFS, "generate/api")
	if getTplErr != nil {
		return getTplErr
	}
	// 清理临时目录
	defer os.RemoveAll(tplDir)

	analysisCfg := &codegen.ApiCfg{
		CommonConfig: codegen.CommonConfig{
			PackageName:       apiGenCfg.PackageName,
			TplDir:            tplDir,
			RootDir:           g.workDir,
			LayerParentDirMap: defaultLayerParentDirMap,
			LayerNameMap:      defaultLayerNameMap,
			LayerPrefixMap:    defaultLayerPrefixMap,
			TplFuncMap: template.FuncMap{
				TplFuncToKebabCase: toKebabCase,
				TplFuncPluralize:   pluralize,
			},
		},
		TargetFilename: apiGenCfg.TargetFilename,
	}
	gen := codegen.NewGenerator()
	analysisRes, analysisErr := gen.AnalysisApiTpl(analysisCfg)
	if analysisErr != nil {
		return fmt.Errorf("analysis api tpl error: %v", analysisErr)
	}

	structName := gutil.SnakeToPascal(gutil.TrimFileExtension(apiGenCfg.TargetFilename))
	structNameLowerCamel := gutil.FirstLetterToLower(structName)
	functionName := gutil.FirstLetterToUpper(apiGenCfg.FunctionName)
	functionNameLowerCamel := gutil.FirstLetterToLower(apiGenCfg.FunctionName)
	// 资源路径统一为 kebab-case 复数（与 module 模板的 restful 风格一致）
	resourcePath := toKebabCase(pluralize(structNameLowerCamel))

	fmt.Printf("[API] Generating API - Method: %s, Path: /%s/%s, Description: %s\n",
		apiGenCfg.HttpMethod, resourcePath, functionNameLowerCamel, apiGenCfg.Description)
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
					ProjectName:     g.appInfo.ProjectName,
					AppName:         g.appInfo.AppName,
					ProjectRootPath: g.appInfo.ProjectRootPath,
					BaseModulePath:  g.appInfo.BaseModulePath,
					AppModuleName:   g.appInfo.AppModuleName,
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
	// 幂等：若该 controller 方法已存在（同名），说明该 endpoint 已生成过，跳过整段生成，
	// 避免 --force 重复追加导致 dto/方法 重复声明。
	if controllerFilepath != "" {
		if data, rerr := os.ReadFile(controllerFilepath); rerr == nil {
			if strings.Contains(string(data), "func (ctr *"+structNameLowerCamel+"Ctr) "+functionName+"(") {
				fmt.Printf("[API] endpoint %s.%s already exists, skip\n", structNameLowerCamel, functionName)
				return nil
			}
		}
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
		routerCallContent := fmt.Sprintf("%sRouter(%s)", structNameLowerCamel, "groups")
		routerEnterFilepath := filepath.Join(g.workDir, "/internal/router/router.go")
		if err := addContentToFuncIdempotent(routerEnterFilepath, "RegisterRouter", routerCallContent); err != nil {
			return fmt.Errorf("new router appendContentToFunc error: %v", err)
		}
		fmt.Printf("[API] Registered new router: %sRouter\n", structNameLowerCamel)
	} else {
		routerCallContent := fmt.Sprintf(`v1RouterGroup.%s("/%s/%s", %sCtr.%s)`, apiGenCfg.HttpMethod, resourcePath, functionNameLowerCamel, structNameLowerCamel, functionName)
		routerEnterFilepath := filepath.Join(g.workDir, fmt.Sprintf("/internal/router/%s.go", gutil.TrimFileExtension(apiGenCfg.PackageName)))
		// 使用 AddContentToFunc 添加到函数末尾，避免注释丢失
		if data, rerr := os.ReadFile(routerEnterFilepath); rerr == nil && strings.Contains(string(data), routerCallContent) {
			// 已存在该路由行，跳过
		} else if err := gast.AddContentToFuncWithLineNumber(routerEnterFilepath, fmt.Sprintf("%sRouter", structNameLowerCamel), routerCallContent, -1); err != nil {
			return fmt.Errorf("appendContentToFunc error: %v", err)
		}
		fmt.Printf("[API] Added route to existing router: %s %s\n", apiGenCfg.HttpMethod, fmt.Sprintf("/%s/%s", resourcePath, functionNameLowerCamel))
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

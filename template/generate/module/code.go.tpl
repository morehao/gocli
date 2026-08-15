package code

import "github.com/morehao/golib/gerror"

const (
    {{.StructName}}CreateError      = 100100
    {{.StructName}}DeleteError      = 100101
    {{.StructName}}UpdateError      = 100102
    {{.StructName}}GetDetailError   = 100103
    {{.StructName}}GetPageListError = 100104
    {{.StructName}}NotExistError    = 100105
)

var {{.StructNameLowerCamel}}ErrorMsgMap = gerror.CodeMsgMap{
    {{.StructName}}CreateError:      "创建{{.Description}}失败",
    {{.StructName}}DeleteError:      "删除{{.Description}}失败",
    {{.StructName}}UpdateError:      "修改{{.Description}}失败",
    {{.StructName}}GetDetailError:   "查看{{.Description}}失败",
    {{.StructName}}GetPageListError: "查看{{.Description}}列表失败",
    {{.StructName}}NotExistError:    "{{.Description}}不存在",
}
/*
 * @Author: morehao morehao@qq.com
 * @Date: 2025-11-08 17:44:59
 * @LastEditors: morehao morehao@qq.com
 * @LastEditTime: 2025-11-08 17:46:54
 * @FilePath: /golib/Users/morehao/Documents/practice/go/go-gin-web/apps/demo/client/httpbingo/httpbingo.go
 * @Description: 这默认设置,请设置`customMade`, 打开koroFileHeader查看配置 进行设置: https://github.com/OBKoro1/koro1FileHeader/wiki/%E9%85%8D%E7%BD%AE
 */
package httpbingo

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/morehao/go-ark-template/demo/config"

	"github.com/morehao/golib/protocol/ghttp"
)

func Get(ctx *gin.Context, req *GetRequest) (*GetResponse, error) {
	cfg := config.Conf
	var res GetResponse

	// 使用新的 ghttp API
	// 方式1: 使用 GetJSON 直接映射到结构体
	params := map[string]string{
		"id": strconv.FormatUint(uint64(req.ID), 10),
	}

	err := cfg.Client.HTTPBingo.GetJSON(ctx, "/get", &res, ghttp.RequestOption{
		RequestBody: params,
	})
	if err != nil {
		return nil, err
	}

	return &res, nil
}

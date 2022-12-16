package main

import (
	"context"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/glog"
	"github.com/gogf/gf/v2/os/gtimer"
	"github.com/gogf/gf/v2/util/gconv"
	"time"
)

func main() {
	var (
		ctx           = gctx.New()
		interval      = 5 * time.Second
		defaultStatus = "已售罄" // 默认状态
		jdlcUrl       = "https://ms.jr.jd.com/gw/generic/pc_bx/h5/m/getPensionItem"
		pushdeerUrl   = "https://api2.pushdeer.com/message/push"
	)

	gtimer.Add(ctx, interval, func(ctx context.Context) {
		// 获取要监听的 skuId 的配置
		skuId, err := g.Cfg().Get(ctx, "skuId")
		if err != nil {
			glog.Error(ctx, "Get SkuId Error:", err.Error())
			return
		}

		// 查询是否开售
		response, err := g.Client().Post(ctx, jdlcUrl, g.Map{
			"reqData": "{\"skuId\":" + "\"" + gconv.String(skuId) + "\"" + ",\"productId\":\"\",\"pageLevel\":\"1\"}",
		})
		if err != nil || response.Response.StatusCode != 200 {
			var errorMessage string
			if len(err.Error()) > 0 {
				errorMessage = err.Error()
			} else {
				errorMessage = response.ReadAllString()
			}
			glog.Error(ctx, "JD API ERROR:", errorMessage)
		}
		j := gjson.New(response.ReadAllString())
		currentStatus := gconv.String(j.Get("resultData.value.buttonStatus.text"))

		if len(currentStatus) == 0 {
			glog.Error(ctx, "Current Status is empty!")
			return
		}
		// 产品开售状态发生变更
		if currentStatus != defaultStatus {
			pushDeerList, err := g.Cfg().Get(ctx, "pushDeerList")
			if err != nil {
				glog.Error(ctx, "Get PushDeer List Error:", err.Error())
				return
			}

			// 发送通知
			for _, pushDeer := range gconv.SliceAny(pushDeerList) {
				_, err := g.Client().Get(ctx, pushdeerUrl, g.Map{
					"pushkey": pushDeer,
					"text":    "【" + currentStatus + "】- 安增益4号 售卖状态已更新",
				})
				if err != nil {
					glog.Error(ctx, "Push Message Error:", err.Error())
					continue
				}
			}
			// 更新 defaultStatus 为当前状态
			defaultStatus = currentStatus
		}
	})

	select {}
}

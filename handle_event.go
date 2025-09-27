package main

import (
	"log"
	"main/mp/core"
)

func defaultEventHandler(ctx *core.Context) {
	log.Printf("收到事件:\n%s\n", ctx.MsgPlaintext)
	eventType := ctx.MixedMsg.EventType
	if eventType == "subscribe" {
		dealSubscribe(ctx)
	} else if eventType != "" {
		responseText(ctx, "对不起，我现在还不能处理 "+string(eventType)+" 事件")
	} else {
		ctx.NoneResponse()
	}
}

func dealSubscribe(ctx *core.Context) {
	log.Printf("收到订阅事件: %s", ctx.MixedMsg.FromUserName)
	responseText(ctx, CONFIG.SubscribeMessage)
}

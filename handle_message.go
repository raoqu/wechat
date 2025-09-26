package main

import (
	"log"
	"main/mp/core"
)

func defaultMsgHandler(ctx *core.Context) {
	log.Printf("收到消息:\n%s\n", ctx.MsgPlaintext)
	msgType := ctx.MixedMsg.MsgType
	if msgType == "text" {
		dealText(ctx)
	} else if msgType != "" {
		responseText(ctx, "对不起，我现在还不能处理 "+string(msgType)+" 消息")
	} else {
		ctx.NoneResponse()
	}
}

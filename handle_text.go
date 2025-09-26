package main

import (
	"log"
	"main/mp/core"
	"main/mp/message/callback/request"
	"main/mp/message/callback/response"
)

func textMsgHandler(ctx *core.Context) {
	msgType := ctx.MixedMsg.MsgType
	if msgType == "text" {
		log.Printf("收到文本消息: %s [from %s]", ctx.MixedMsg.Content, ctx.MixedMsg.FromUserName)
		dealText(ctx)
	} else {
		log.Printf("收到文本消息:\n%s\n", ctx.MsgPlaintext)
		responseText(ctx, "这不是文本消息")
	}
}

func responseText(ctx *core.Context, message string) {
	msg := request.GetText(ctx.MixedMsg)
	resp := response.NewText(msg.FromUserName, msg.ToUserName, msg.CreateTime, message)
	ctx.AESResponse(resp, 0, "", nil) // aes密文回复
}

func dealText(ctx *core.Context) {
	msg := request.GetText(ctx.MixedMsg)
	responseText(ctx, msg.Content)
}

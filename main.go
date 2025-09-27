package main

import (
	"log"
	"net/http"

	"main/mp/core"
	"main/mp/menu"
	"main/mp/message/callback/request"
	"main/mp/message/callback/response"

	"github.com/raoqu/gojs"
)

var (
	// 下面两个变量不一定非要作为全局变量, 根据自己的场景来选择.
	msgHandler core.Handler
	msgServer  *core.Server
	GOJS       *gojs.GoJSInstance
)

func init() {
	mux := core.NewServeMux()
	mux.DefaultMsgHandleFunc(defaultMsgHandler)
	mux.DefaultEventHandleFunc(defaultEventHandler)
	mux.MsgHandleFunc(request.MsgTypeText, textMsgHandler)
	mux.EventHandleFunc(menu.EventTypeClick, menuClickEventHandler)

	msgHandler = mux
	msgServer = core.NewServer(CONFIG.OriID, CONFIG.AppID, CONFIG.Token, CONFIG.EncodedAESKey, msgHandler, nil)
}

func menuClickEventHandler(ctx *core.Context) {
	log.Printf("收到菜单 click 事件:\n%s\n", ctx.MsgPlaintext)

	event := menu.GetClickEvent(ctx.MixedMsg)
	resp := response.NewText(event.FromUserName, event.ToUserName, event.CreateTime, "收到 click 类型的事件")
	//ctx.RawResponse(resp) // 明文回复
	ctx.AESResponse(resp, 0, "", nil) // aes密文回复
}

func init() {
	http.HandleFunc("/wx_notify", wxCallbackHandler)

	// WebApp 登录相关
	http.HandleFunc("/auth/url", AuthURL)           // 返回微信授权链接
	http.HandleFunc("/auth/callback", AuthCallback) // 微信回调
	http.HandleFunc("/auth/me", Me)                 // 已登录用户信息示例

}

// wxCallbackHandler 是处理回调请求的 http handler.
//  1. 不同的 web 框架有不同的实现
//  2. 一般一个 handler 处理一个公众号的回调请求(当然也可以处理多个, 这里我只处理一个)
func wxCallbackHandler(w http.ResponseWriter, r *http.Request) {
	msgServer.ServeHTTP(w, r, nil)
}

func main() {
	jsConfig := gojs.LoadConfig("gojs.yaml")
	GOJS := gojs.CreateInstance(jsConfig)
	GOJS.Run()
	go func() {
		http.ListenAndServe(CONFIG.ServerAddr, nil)
	}()
	select {}
}

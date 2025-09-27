package main

type WxAppConfig struct {
	AppID         string
	AppSecret     string
	OriID         string
	Token         string
	EncodedAESKey string

	// WebApp 相关(可选)
	RedirectURI  string
	JWTSecret    string // 站点的 JWT 密钥（生成自己的随机值）
	CookieDomain string

	// 其他
	SubscribeMessage string // 订阅事件回复
	ServerAddr       string // 微信服务器地址
}

var DEFAULT_CONFIG = WxAppConfig{
	AppID:         "appId",
	AppSecret:     "appSecret",
	OriID:         "gh_...",
	Token:         "token",
	EncodedAESKey: "encodedAESKey",

	// WebApp 相关(可选)
	RedirectURI:  "redirectUri",
	JWTSecret:    "jwtSecret",
	CookieDomain: "cookieDomain",

	SubscribeMessage: "感谢您的关注！",
	ServerAddr:       ":7666",
}

var CONFIG = DEFAULT_CONFIG

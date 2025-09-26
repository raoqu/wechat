package main

import (
 	"encoding/json"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// 内存态 state 存储（生产建议用 Redis）
var stateStore = NewStateStore(5 * time.Minute)

type StateStore struct {
	m   map[string]time.Time
	ttl time.Duration
}

func NewStateStore(ttl time.Duration) *StateStore {
	return &StateStore{m: map[string]time.Time{}, ttl: ttl}
}
func (s *StateStore) Put(k string) { s.m[k] = time.Now().Add(s.ttl) }
func (s *StateStore) Verify(k string) bool {
	exp, ok := s.m[k]
	if !ok {
		return false
	}
	delete(s.m, k)
	return time.Now().Before(exp)
}

func randomState() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

// GET /auth/url?redirect=<回跳到业务页的URL>&scope=base|userinfo
func AuthURL(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	scope := q.Get("scope") // base: snsapi_base, userinfo: snsapi_userinfo
	if scope == "" { scope = "base" }
	redirectTo := q.Get("redirect")
	if !strings.HasPrefix(redirectTo, "/") && !strings.HasPrefix(redirectTo, "http") {
		redirectTo = "/"
	}

	realScope := map[string]string{
		"base":     "snsapi_base",
		"userinfo": "snsapi_userinfo",
	}[scope]

	state := randomState()
	stateStore.Put(state)

	authURL := fmt.Sprintf(
		"https://open.weixin.qq.com/connect/oauth2/authorize?appid=%s&redirect_uri=%s&response_type=code&scope=%s&state=%s#wechat_redirect",
		url.QueryEscape(wxAppId),
		url.QueryEscape(RedirectURI+"?redirect="+url.QueryEscape(redirectTo)),
		url.QueryEscape(realScope),
		url.QueryEscape(state),
	)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(map[string]string{"authorize_url": authURL})
}

// GET /auth/callback?code=xxx&state=yyy&redirect=/xxx
func AuthCallback(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	code := q.Get("code")
	state := q.Get("state")
	redirectTo := q.Get("redirect")
	if redirectTo == "" { redirectTo = "/" }

	if code == "" || !stateStore.Verify(state) {
		http.Error(w, "invalid code/state", http.StatusBadRequest)
		return
	}

	// 1) 用 code 换取 access_token / openid
	tokenAPI := fmt.Sprintf(
		"https://api.weixin.qq.com/sns/oauth2/access_token?appid=%s&secret=%s&code=%s&grant_type=authorization_code",
		url.QueryEscape(wxAppId),
		url.QueryEscape(wxAppSecret),
		url.QueryEscape(code),
	)
	resp, err := http.Get(tokenAPI)
	if err != nil {
		http.Error(w, fmt.Sprintf("weixin token api err: %v", err), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	var tk struct {
		AccessToken  string `json:"access_token"`
		ExpiresIn    int64  `json:"expires_in"`
		RefreshToken string `json:"refresh_token"`
		Openid       string `json:"openid"`
		Scope        string `json:"scope"`
		Unionid      string `json:"unionid"`
		ErrCode      int    `json:"errcode"`
		ErrMsg       string `json:"errmsg"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tk); err != nil || tk.ErrCode != 0 || tk.Openid == "" {
		http.Error(w, fmt.Sprintf("decode token resp err: %v, wxerr=%d %s", err, tk.ErrCode, tk.ErrMsg), http.StatusBadGateway)
		return
	}

	// 2) 如果 scope 包含 snsapi_userinfo，可拉取用户资料
	var profile *WXUserInfo
	if strings.Contains(tk.Scope, "snsapi_userinfo") {
		uiAPI := fmt.Sprintf("https://api.weixin.qq.com/sns/userinfo?access_token=%s&openid=%s&lang=zh_CN",
			url.QueryEscape(tk.AccessToken), url.QueryEscape(tk.Openid))
		r2, err := http.Get(uiAPI)
		if err == nil {
			defer r2.Body.Close()
			var ui WXUserInfo
			if json.NewDecoder(r2.Body).Decode(&ui) == nil && ui.Openid != "" {
				profile = &ui
			}
		}
	}

	// 3) 本地用户落库/查询（这里用内存模拟）
	u := UpsertUser(tk.Openid, tk.Unionid, profile)

	// 4) 签发站点 JWT
	token, err := SignJWT(u.ID, u.OpenID, u.UnionID)
	if err != nil {
		http.Error(w, fmt.Sprintf("sign jwt err: %v", err), http.StatusInternalServerError)
		return
	}

	// 5) 设置 Cookie 并跳回业务页
	http.SetCookie(w, &http.Cookie{
		Name:     "site_jwt",
		Value:    token,
		Path:     "/",
		Domain:   CookieDomain,
		Secure:   true,
		HttpOnly: true,
		MaxAge:   3600 * 24 * 30,
		SameSite: http.SameSiteLaxMode,
	})
	http.Redirect(w, r, redirectTo, http.StatusFound)
}

type WXUserInfo struct {
	Openid     string `json:"openid"`
	Nickname   string `json:"nickname"`
	Sex        int    `json:"sex"`
	Province   string `json:"province"`
	City       string `json:"city"`
	Country    string `json:"country"`
	HeadImgURL string `json:"headimgurl"`
	UnionID    string `json:"unionid"`
	// ... 其他字段按需扩展
}

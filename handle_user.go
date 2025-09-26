package main

import "sync"

type User struct {
	ID      int64
	OpenID  string
	UnionID string
	Nick    string
	Avatar  string
}

var (
	usersMu sync.Mutex
	users         = map[string]*User{} // key: openid
	uidSeq  int64 = 1000
)

func UpsertUser(openid, unionid string, ui *WXUserInfo) *User {
	usersMu.Lock()
	defer usersMu.Unlock()

	if u, ok := users[openid]; ok {
		// 更新资料
		if ui != nil {
			u.Nick = ui.Nickname
			u.Avatar = ui.HeadImgURL
			if ui.UnionID != "" {
				u.UnionID = ui.UnionID
			}
		}
		return u
	}
	uidSeq++
	u := &User{
		ID:      uidSeq,
		OpenID:  openid,
		UnionID: unionid,
	}
	if ui != nil {
		u.Nick = ui.Nickname
		u.Avatar = ui.HeadImgURL
		if ui.UnionID != "" {
			u.UnionID = ui.UnionID
		}
	}
	users[openid] = u
	return u
}

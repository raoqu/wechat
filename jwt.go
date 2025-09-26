package main

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID  int64  `json:"uid"`
	OpenID  string `json:"openid"`
	UnionID string `json:"unionid,omitempty"`
	jwt.RegisteredClaims
}

type ctxKey string

const claimsKey ctxKey = "claims"

func SignJWT(uid int64, openid, unionid string) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:  uid,
		OpenID:  openid,
		UnionID: unionid,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "wechat-h5-login",
			Subject:   openid,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(30 * 24 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(JWTSecret))
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Try Cookie first
		if c, err := r.Cookie("site_jwt"); err == nil && c.Value != "" {
			if claims, ok := parseJWT(c.Value); ok {
				ctx := context.WithValue(r.Context(), claimsKey, claims)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
		}
		// Then Authorization: Bearer <token>
		if ah := r.Header.Get("Authorization"); len(ah) > 7 && ah[:7] == "Bearer " {
			if claims, ok := parseJWT(ah[7:]); ok {
				ctx := context.WithValue(r.Context(), claimsKey, claims)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
		}
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	})
}

func parseJWT(token string) (*Claims, bool) {
	t, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(JWTSecret), nil
	})
	if err != nil {
		return nil, false
	}
	if cl, ok := t.Claims.(*Claims); ok && t.Valid {
		return cl, true
	}
	return nil, false
}

func SetAuthCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "site_jwt",
		Value:    token,
		Path:     "/",
		Domain:   CookieDomain,
		Secure:   true, // 生产用 https
		HttpOnly: true,
		MaxAge:   3600 * 24 * 30, // 30 天
		SameSite: http.SameSiteLaxMode,
	})
}

func Me(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if claims, ok := r.Context().Value(claimsKey).(*Claims); ok && claims != nil {
		_ = json.NewEncoder(w).Encode(map[string]*Claims{"me": claims})
		return
	}
	w.WriteHeader(http.StatusUnauthorized)
	_, _ = w.Write([]byte(`{"error":"unauthorized"}`))
}

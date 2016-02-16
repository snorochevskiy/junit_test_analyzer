package main

import (
	"crypto/rand"
	"encoding/base64"
	"io"
	"net/http"
	"time"
)

const SID_COOKIE_NAME = "SID"

var SESSION_MANAGER SessionManagerService = SessionManagerService{sessionMap: make(map[string]*Session)}

func init() {

}

type Session struct {
	Sid     string
	User    *UserEntity
	Created time.Time
}

func (session *Session) IsLoggedIn() bool {
	return session.User != nil
}

func (session *Session) GetUserRenderInfo() *UserRenderInfo {
	u := new(UserRenderInfo)
	if session.User == nil {
		u.LoggedIn = false
	} else {
		u.LoggedIn = true
		u.Details = session.User
	}
	return u
}

type SessionManagerService struct {
	sessionMap map[string]*Session
}

func (sessionManager *SessionManagerService) InitSession(w http.ResponseWriter, userInfo *UserEntity) {
	generatedSid := sessionManager.generateSid()
	session := &Session{Sid: generatedSid}
	session.User = userInfo

	sessionManager.sessionMap[generatedSid] = session

	cookie := &http.Cookie{Name: SID_COOKIE_NAME, Value: session.Sid, MaxAge: 0}
	http.SetCookie(w, cookie)
}

func (sessionManager *SessionManagerService) ClearSession(r *http.Request, w http.ResponseWriter) {
	cookie, err := r.Cookie(SID_COOKIE_NAME)
	if err != nil {
		return
	}
	delete(sessionManager.sessionMap, cookie.Value)

	http.SetCookie(w, &http.Cookie{Name: SID_COOKIE_NAME, Value: "", Path: "/", MaxAge: -1})
}

func (manager *SessionManagerService) generateSid() string {
	b := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return ""
	}
	return base64.URLEncoding.EncodeToString(b)
}

func (manager *SessionManagerService) GetSessionForRequest(r *http.Request) *Session {

	cookie, err := r.Cookie(SID_COOKIE_NAME)
	if err != nil {
		return new(Session)
	}
	sid := cookie.Value

	return manager.sessionMap[sid]
}

func (session *Session) SetCookie(w http.ResponseWriter) {
	cookie := &http.Cookie{Name: SID_COOKIE_NAME, Value: session.Sid, MaxAge: 0}
	http.SetCookie(w, cookie)
}

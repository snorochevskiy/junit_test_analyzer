package main

import (
	"net/http"
)

type ValidationError struct {
	Message string
}

type IValidator interface {
	Validate(*Context) *ValidationError
}

type ISessionProvider interface {
	GetSessionForRequest(*http.Request) *Session
}

type User struct {
	UserId      int64
	Login       string
	Password    string
	Permissions []string
}

type Context struct {
	User *UserEntity
}

type Wrapper struct {
	SessionProvider ISessionProvider
	Validators      []IValidator

	HadleFunc func(*Context) interface{}
}

type Renderer struct {
}

func (wrapper *Wrapper) Wrap() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		context := new(Context)

		session := wrapper.SessionProvider.GetSessionForRequest(r)
		if session != nil {
			context.User = session.User
		}

		for _, validator := range wrapper.Validators {
			if validErr := validator.Validate(context); validErr != nil {
				http.Error(w, validErr.Message, http.StatusBadRequest)
			}
		}

	}
}

type AuthWrapper struct {
	SessionProvider     ISessionProvider
	SessionIdCookieName string
}

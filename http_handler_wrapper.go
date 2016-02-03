package main

import (
	"net/http"
)

const (
	HND_SUCCESS = 0
	HND_FAIL    = 1
)

type ValidationError struct {
	Message string
}

type HndStatus struct {
	Code    int
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

	HadleFunc func(*Context) (interface{}, HndStatus)

	ResultRenderer *GoTemplateRenderer
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

		result, status := wrapper.HadleFunc(context)
		if status.Code != HND_FAIL {

		} else {
			wrapper.ResultRenderer.Render(w, result)
		}

	}
}

type GoTemplateRenderer struct {
	TemplateName string
}

func (r *GoTemplateRenderer) Render(w http.ResponseWriter, data interface{}) {
	RenderInCommonTemplate(w, data, r.TemplateName)
}

type ErrorsRenderer struct {
}

func (er *ErrorsRenderer) Render(w http.ResponseWriter, he HndStatus) {
	http.Error(w, he.Message, http.StatusInternalServerError)
}

type AuthWrapper struct {
	SessionProvider     ISessionProvider
	SessionIdCookieName string
}

func CreateDeleteHandlerWrapper() {

}

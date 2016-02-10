package main

import (
	"net/http"
	"strconv"
)

const (
	HND_SUCCESS = 0
	HND_FAIL    = 1
)

type ParamType int
type ParamSource int

const (
	PARAM_TYPE_STRING ParamType = iota
	PARAM_TYPE_INT
)

type ValidationError struct {
	Message string
}

type HndError struct {
	Code    int
	Message string
	Type    string
}

func (he *HndError) Error() string {
	return he.Message
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
	Req    *http.Request
	Resp   http.ResponseWriter
	User   *UserEntity
	Params map[string]interface{}
}

type Param struct {
	Name      string
	Mandatory bool
	Type      ParamType
}

type Wrapper struct {
	SessionProvider ISessionProvider
	ExpectedParams  []Param
	Validators      []IValidator

	HadleFunc func(*Context) (interface{}, error)

	SuccessRenderer IRenderer
	ErrorRenderer   IRenderer
}

func (wrapper *Wrapper) Wrap() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		context := new(Context)
		context.Req = r
		context.Resp = w
		context.Params = make(map[string]interface{})

		if wrapper.SessionProvider != nil {
			session := wrapper.SessionProvider.GetSessionForRequest(r)
			if session != nil {
				context.User = session.User
			}
		}

		for _, expPar := range wrapper.ExpectedParams {
			p := r.URL.Query().Get(expPar.Name)
			if p == "" {
				if expPar.Mandatory {
					http.Error(w, expPar.Name+" parameter not found", http.StatusBadRequest)
					return
				} else {
					continue
				}
			}
			if expPar.Type == PARAM_TYPE_STRING {
				context.Params[expPar.Name] = p
			} else if expPar.Type == PARAM_TYPE_INT {
				num, err := strconv.Atoi(p)
				if err != nil {
					http.Error(w, expPar.Name+" is not a number", http.StatusBadRequest)
					return
				}
				context.Params[expPar.Name] = num
			}
		}

		for _, validator := range wrapper.Validators {
			if validErr := validator.Validate(context); validErr != nil {
				http.Error(w, validErr.Message, http.StatusBadRequest)
			}
		}

		result, err := wrapper.HadleFunc(context)
		if err != nil {
			wrapper.ErrorRenderer.Render(context, result)
		} else {
			wrapper.SuccessRenderer.Render(context, result)
		}
	}
}

type IRenderer interface {
	Render(c *Context, data interface{})
}

type GoTemplateRenderer struct {
	TemplateName string
}

func (r *GoTemplateRenderer) Render(c *Context, data interface{}) {
	if renderErr := RenderInCommonTemplate(c.Resp, data, r.TemplateName); renderErr != nil {
		http.Error(c.Resp, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

type ErrorsRenderer struct {
}

func (er *ErrorsRenderer) Render(w http.ResponseWriter, he HndError) {
	http.Error(w, he.Message, http.StatusInternalServerError)
}

type AuthWrapper struct {
	SessionProvider     ISessionProvider
	SessionIdCookieName string
}

func CreateDeleteHandlerWrapper() {

}

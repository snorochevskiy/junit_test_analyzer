package main

import (
	"net/http"
	"strings"
)

type PathElementType int

const (
	PATH_ELEMENT_EXACT PathElementType = iota
	PATH_ELEMENT_VARIABLE
)

type RoutedHandler struct {
	Routes []Route
}

type HttpContext struct {
	Session *Session
	Req     *http.Request
	Resp    http.ResponseWriter
}

type Route struct {
	Path    []PathElement
	Handler func(*HttpContext)
}

type PathElement struct {
	Val  string
	Type PathElementType
}

func (mh *RoutedHandler) AddRoute(urlPattern string, handler func(*HttpContext)) error {
	if mh.Routes == nil {
		mh.Routes = make([]Route, 0, 10)
	}

	parts := strings.Split(urlPattern, "/")
	route := Route{Path: make([]PathElement, 0, 3), Handler: handler}

	for _, v := range parts {
		if strings.HasSuffix(v, ":") {
			route.Path = append(route.Path, PathElement{Val: v[1:], Type: PATH_ELEMENT_VARIABLE})
		} else {
			route.Path = append(route.Path, PathElement{Val: v, Type: PATH_ELEMENT_EXACT})
		}
	}
	mh.Routes = append(mh.Routes, route)

	return nil
}

func (mh *RoutedHandler) FindMatchingRoute(url string) *Route {

	urlParts := strings.Split(url, "/")

Loop:
	for _, v := range mh.Routes {
		if len(urlParts) != len(v.Path) {
			continue
		}
		for ind := 0; ind < len(urlParts); ind++ {
			if urlParts[ind] != v.Path[ind].Val && v.Path[ind].Type != PATH_ELEMENT_VARIABLE {
				continue Loop
			}
		}
		return &v
	}
	return nil
}

func (mh RoutedHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	url := r.URL.EscapedPath()

	// TODO: fix favicon issue for apis
	if strings.Contains(url, "favicon") {
		return
	}

	route := mh.FindMatchingRoute(url)
	if route == nil {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	}

	session := SESSION_MANAGER.GetSessionForRequest(r)
	context := HttpContext{
		Session: session,
		Req:     r,
		Resp:    w,
	}

	route.Handler(&context)
}

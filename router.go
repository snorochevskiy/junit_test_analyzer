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

func (p PathElementType) String() string {
	switch p {
	case PATH_ELEMENT_EXACT:
		return "PATH_ELEMENT_EXACT"
	case PATH_ELEMENT_VARIABLE:
		return "PATH_ELEMENT_VARIABLE"
	}
	return "UNKNOW"
}

type RoutedHandler struct {
	Routes []Route
}

type HttpContext struct {
	Session    *Session
	Req        *http.Request
	Resp       http.ResponseWriter
	PathParams map[string]string
}

type Route struct {
	Path    []PathElement
	Handler func(*HttpContext)
}

type MatchingRoute struct {
	PathParams map[string]string
	Handler    func(*HttpContext)
}

func (mr *MatchingRoute) AddPathParam(name string, value string) {
	if mr.PathParams == nil {
		mr.PathParams = make(map[string]string)
	}
	mr.PathParams[name] = value
}

type PathElement struct {
	Val  string
	Type PathElementType
}

func (pe PathElement) String() string {
	return "PathElement[Val=" + pe.Val + ",Type=" + pe.Type.String() + "]"
}

func (mh *RoutedHandler) AddRoute(urlPattern string, handler func(*HttpContext)) error {
	if mh.Routes == nil {
		mh.Routes = make([]Route, 0, 10)
	}

	parts := strings.Split(urlPattern, "/")
	route := Route{Path: make([]PathElement, 0, 3), Handler: handler}

	for _, v := range parts {
		if strings.HasPrefix(v, ":") {
			route.Path = append(route.Path, PathElement{Val: v[1:], Type: PATH_ELEMENT_VARIABLE})
		} else {
			route.Path = append(route.Path, PathElement{Val: v, Type: PATH_ELEMENT_EXACT})
		}
	}
	mh.Routes = append(mh.Routes, route)

	return nil
}

func (mh *RoutedHandler) FindMatchingRoute(url string) *MatchingRoute {

	urlParts := strings.Split(url, "/")
	matchingRoute := new(MatchingRoute)

Loop:
	for _, v := range mh.Routes {

		if len(urlParts) != len(v.Path) {
			continue
		}

		for ind := 0; ind < len(urlParts); ind++ {
			if urlParts[ind] != v.Path[ind].Val && v.Path[ind].Type != PATH_ELEMENT_VARIABLE {
				continue Loop
			}
			if v.Path[ind].Type == PATH_ELEMENT_VARIABLE {
				matchingRoute.AddPathParam(v.Path[ind].Val, urlParts[ind])
			}
		}
		matchingRoute.Handler = v.Handler
		return matchingRoute
	}
	return nil
}

func (mh RoutedHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	url := r.URL.EscapedPath()

	// TODO: fix favicon issue for apis
	if strings.Contains(url, "favicon") {
		return
	}

	matchingRoute := mh.FindMatchingRoute(url)
	if matchingRoute == nil {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	session := SESSION_MANAGER.GetSessionForRequest(r)
	context := HttpContext{
		Session:    session,
		Req:        r,
		Resp:       w,
		PathParams: matchingRoute.PathParams,
	}

	matchingRoute.Handler(&context)
}

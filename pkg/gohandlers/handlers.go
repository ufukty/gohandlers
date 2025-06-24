package gohandlers

import "net/http"

type HandlerInfo struct {
	Method string
	Path   string
	Ref    http.HandlerFunc
}

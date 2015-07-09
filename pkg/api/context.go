package api

import (
	"github.com/qorio/omni/auth"
	"net/http"
)

type Context interface {
	UserId() string
	UrlParameter(string) string
}

type CreateContextFunc func(auth.Context, *http.Request) Context

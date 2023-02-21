package binding

import (
	"net/http"
	"strings"
)

// MIME
const (
	MIMEJSON = "application/json"
)

// Binding interface
type Binding interface {
	Name() string
	Bind(*http.Request, interface{}) error
}

var (
	// JSON bind
	JSON = jsonBinding{}
	// QUERY bind
	QUERY = queryBinding{}
)

// Default bind
func Default(method, contentType string) Binding {
	if method == "GET" || method == "DELETE" {
		return QUERY
	}

	contentType = stripContentType(contentType)
	switch contentType {
	case MIMEJSON:
		return JSON
	default: // MIMEForm, MIMEPOSTForm, MIMEMultipartPOSTForm
		return nil
	}
}

func stripContentType(contentType string) string {
	return strings.TrimSpace(strings.Split(contentType, ";")[0])
}

// Bind bind
func Bind(r *http.Request, v interface{}) error {
	b := Default(r.Method, r.Header.Get("Content-Type"))
	return mustBindWith(r, v, b)
}

func mustBindWith(r *http.Request, v interface{}, b Binding) (err error) {
	err = b.Bind(r, v)
	return
}

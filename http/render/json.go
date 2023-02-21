package render

import (
	"errors"
	"net/http"

	"github.com/unrolled/render"
)

var (
	r *render.Render
)

func init() {
	r = render.New()
}

// JSON render
func JSON(w http.ResponseWriter, obj interface{}) {
	r.JSON(w, http.StatusOK, obj)
}

// ErrResp struct
type ErrResp struct {
	Err            error `json:"-"` // low-level runtime error
	HTTPStatusCode int   `json:"-"` // http response status code

	Detail    string `json:"detail"`          // user-level status message
	ErrorText string `json:"error,omitempty"` // application-level error message, for debugging
}

// JSONErr render
func JSONErr(w http.ResponseWriter, code int, errs ...error) {
	err := errors.New(http.StatusText(code))
	if errs != nil {
		err = errs[0]
	}
	v := &ErrResp{
		Err:            err,
		HTTPStatusCode: code,
		Detail:         err.Error(),
	}
	r.JSON(w, v.HTTPStatusCode, v)
}

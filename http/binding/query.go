package binding

import (
	"net/http"

	"github.com/gorilla/schema"
)

var (
	decoder *schema.Decoder
)

func init() {
	decoder = schema.NewDecoder()
	decoder.IgnoreUnknownKeys(true)
}

type queryBinding struct{}

func (q queryBinding) Name() string {
	return "url query"
}

func (q queryBinding) Bind(r *http.Request, v interface{}) (err error) {
	err = decoder.Decode(v, r.URL.Query())
	return
}

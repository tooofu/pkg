package binding

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
)

type jsonBinding struct{}

func (j jsonBinding) Name() string {
	return "json"
}

func (j jsonBinding) Bind(r *http.Request, v interface{}) (err error) {
	defer io.Copy(ioutil.Discard, r.Body)
	dec := json.NewDecoder(r.Body)
	if err = dec.Decode(v); err != nil {
		return errors.WithStack(err)
	}
	return validate(v)
}

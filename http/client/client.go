package client

import (
	"context"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/opentracing-contrib/go-stdlib/nethttp"
	"github.com/opentracing/opentracing-go"
)

func New(maxIdle int, idleTimeout time.Duration, compress, verify bool) *http.Client {
	return &http.Client{
		// Transport: &http.Transport{
		//     MaxIdleConns:       maxIdle,
		//     IdleConnTimeout:    idleTimeout,
		//     DisableCompression: compress,
		//     TLSClientConfig: &tls.Config{
		//         InsecureSkipVerify: verify,
		//     },
		// },
		Transport: &nethttp.Transport{},
	}
}

func Do(c context.Context, cli *http.Client, r *http.Request, tr opentracing.Tracer) (bs []byte, err error) {
	var resp *http.Response

	r = r.WithContext(c)

	r, ht := nethttp.TraceRequest(tr, r)
	defer ht.Finish()

	if resp, err = cli.Do(r); err != nil {
		return
	}
	defer resp.Body.Close()

	bs, err = ioutil.ReadAll(resp.Body)
	return
}

package http

import (
	"crypto/tls"
	nethttp "net/http"
	"time"
)

func WithUrl(url string) Option {
	return optionFunc(func(r *Parameter) {
		r.url = url
	})
}

func WithMethod(method Method) Option {
	return optionFunc(func(r *Parameter) {
		r.method = method
	})
}

func WithTimeout(timeout time.Duration) Option {
	return optionFunc(func(r *Parameter) {
		r.timeout = timeout
	})
}

func WithQuery(query map[string]interface{}) Option {
	return optionFunc(func(r *Parameter) {
		for key, val := range query {
			r.query[key] = val
		}
	})
}

func WithBody(body interface{}) Option {
	return optionFunc(func(r *Parameter) {
		r.body = body
	})
}

func WithHeader(headers map[string]string) Option {
	return optionFunc(func(r *Parameter) {
		for key, val := range headers {
			r.header[key] = val
		}
	})
}

func WithReqDeal(deal func(r *Parameter) error) Option {
	return optionFunc(func(r *Parameter) {
		r.reqDeal = append(r.reqDeal, deal)
	})
}

func WithTLSClientConfig(tLSClientConfig *tls.Config) Option {
	return optionFunc(func(r *Parameter) {
		r.tLSClientConfig = tLSClientConfig
	})
}

func WithDisableKeepAlives(disableKeepAlives bool) Option {
	return optionFunc(func(r *Parameter) {
		r.disableKeepAlives = disableKeepAlives
	})
}

func WithRespDeal(deal func(response *nethttp.Response) error) Option {
	return optionFunc(func(r *Parameter) {
		r.respDeal = append(r.respDeal, deal)
	})
}

func WithLog(log ILogger) Option {
	return optionFunc(func(r *Parameter) {
		r.log = log
	})
}

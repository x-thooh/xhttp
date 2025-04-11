package http

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/x-thooh/utils/transform"
	"io"
	nh "net/http"
	nu "net/url"
	"reflect"
	"sync"
	"time"
)

type (
	IHttper interface {
		// Get get
		Get(ctx context.Context, url string, result interface{}, opts ...Option) error
		// Post post
		Post(ctx context.Context, url string, result interface{}, opts ...Option) error
	}
	// Parameter 参数
	Parameter struct {
		// url
		url string
		// 请求方式
		method Method
		// 超时时间
		timeout time.Duration
		// header
		header map[string]string
		// query
		query map[string]interface{}
		// body
		body interface{}
		// reqDeal
		reqDeal []func(r *Parameter) error
		// reader
		bodyReader io.Reader
		// tls
		tLSClientConfig *tls.Config
		// disableKeepAlives
		disableKeepAlives bool
		// log
		log ILogger

		// 返回值
		response *nh.Response
		respDeal []func(r *nh.Response) error
	}
	Option interface {
		apply(*Parameter)
	}
	optionFunc func(*Parameter)
	http       struct {
		*Parameter
	}
)

func (f optionFunc) apply(o *Parameter) {
	f(o)
}

func (p *Parameter) setBody(body io.Reader) {
	p.bodyReader = body
}

var hp = sync.Pool{
	New: func() interface{} {
		return NewHttp()
	},
}

func NewHttp() IHttper {
	h := &http{
		Parameter: &Parameter{
			method:  MethodGet,
			timeout: DefaultTimeOut,
			header: map[string]string{
				"Content-Type": "application/json",
			},
			query: map[string]interface{}{},
			tLSClientConfig: &tls.Config{
				InsecureSkipVerify: false,
			},
			reqDeal: []func(r *Parameter) error{
				func(r *Parameter) error {
					// 组装query
					if len(r.query) > 0 {
						params := nu.Values{}
						netUrl, err := nu.Parse(r.url)
						if err != nil {
							return fmt.Errorf("get json ma err, query: %s, %w", r.url, err)
						}
						for key, value := range r.query {
							params.Add(key, transform.StrVal(value))
						}
						netUrl.RawQuery = params.Encode()
						r.url = netUrl.String()
						if r.log != nil {
							r.log.Println("Get url", r.url)
						}
					}

					// 组装body
					if r.body != nil {
						data, err := json.Marshal(r.body)
						if nil != err {
							return fmt.Errorf("json ma err, query: %v, %w", r, err)
						}
						r.setBody(bytes.NewBuffer(data))
						if r.log != nil {
							r.log.Println("Put url", r.url, string(data))
						}
					}
					return nil
				},
			},
		},
	}
	return h
}

func (r *http) withOpt(opts ...Option) error {
	for _, o := range opts {
		o.apply(r.Parameter)
	}
	return nil
}

func (r *http) request(ctx context.Context, url string, result interface{}, opts ...Option) (err error) {
	opts = append([]Option{WithUrl(url)}, opts...)
	// 可选参数
	if err = r.withOpt(opts...); nil != err {
		return fmt.Errorf("request withOpt err, opts: %v, %w", opts, err)
	}

	// 预处理
	for _, preDeal := range r.reqDeal {
		if err = preDeal(r.Parameter); nil != err {
			return fmt.Errorf("request callback err, r: %v, %w", r, err)
		}
	}

	// 组装request
	req, err := nh.NewRequestWithContext(ctx, string(r.method), r.url, r.bodyReader)
	if nil != err {
		return fmt.Errorf("request NewRequestWithContext err, url: %s, bodyReader: %s, %w", r.url, r.bodyReader, err)
	}

	// 组装header
	for key, value := range r.header {
		req.Header.Set(key, value)
	}

	// 发送请求
	client := &nh.Client{Transport: &nh.Transport{
		DisableKeepAlives: r.disableKeepAlives,
		TLSClientConfig:   r.tLSClientConfig,
	}, Timeout: r.timeout}
	resp, err := client.Do(req)
	if nil != err {
		return fmt.Errorf("request do err, query: %v, %w", req, err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); nil != closeErr {
			errStr := ""
			if err != nil {
				errStr = fmt.Sprintf("(%s)", err.Error())
			}
			err = fmt.Errorf("resp bodyReader close err, %v %w", errStr, closeErr)
		}
	}()

	var bodyByte []byte
	if len(r.respDeal) > 0 {
		// 完整Response
		*r.response = *resp
		// 读取请求
		if bodyByte, err = io.ReadAll(resp.Body); nil != err {
			return fmt.Errorf("request read err, bodyByte: %v, %w", bodyByte, err)
		}
		r.response.Body = io.NopCloser(bytes.NewBuffer(bodyByte))
	}

	for _, respDeal := range r.respDeal {
		if err = respDeal(resp); err != nil {
			return err
		}
	}

	// 不需要解析返回值
	if result == nil {
		_, err = io.Copy(io.Discard, resp.Body)
		return
	}

	// 读取请求
	if len(bodyByte) == 0 {
		if bodyByte, err = io.ReadAll(resp.Body); nil != err {
			return fmt.Errorf("request read err, bodyByte: %v, %w", bodyByte, err)
		}
	}

	// 没有内容
	if len(bodyByte) == 0 {
		return
	}

	// 按照JSON解析返回值
	if json.Valid(bodyByte) {
		if err = json.Unmarshal(bodyByte, &result); nil != err {
			return fmt.Errorf("request json un err, result: %v, %w", result, err)
		}
		return
	}

	// 按照字符串解析返回值
	rv := reflect.ValueOf(result)
	if rv.Kind() != reflect.Ptr {
		return errors.New("result must be a pointer")
	}
	rvv := rv.Elem()
	if rvv.Kind() != reflect.String {
		return errors.New("result must be a string")
	}
	if !rvv.CanSet() {
		return errors.New("result can not set")
	}
	rvv.SetString(string(bodyByte))
	return
}

func (r *http) Get(ctx context.Context, url string, result interface{}, opts ...Option) error {
	// withMethod, WithReqDeal
	return r.request(ctx, url, result, append(opts, WithMethod(MethodGet))...)
}

func (r *http) Post(ctx context.Context, url string, result interface{}, opts ...Option) error {
	// withMethod, WithBody
	return r.request(ctx, url, result, append(opts, WithMethod(MethodPost))...)
}

func Get(ctx context.Context, url string, result interface{}, opts ...Option) error {
	httper := hp.Get().(IHttper)
	defer hp.Put(httper)
	return NewHttp().Get(ctx, url, result, opts...)
}

func Post(ctx context.Context, url string, result interface{}, opts ...Option) error {
	httper := hp.Get().(IHttper)
	defer hp.Put(httper)
	return httper.Post(ctx, url, result, opts...)
}

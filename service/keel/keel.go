package keel

import (
	"context"
	"crypto/rand"
	"io"
	"math/big"
	"net/http"

	"github.com/tkeel-io/tkeel"
	"github.com/tkeel-io/tkeel/keel"
	"github.com/tkeel-io/tkeel/logger"
	"github.com/tkeel-io/tkeel/openapi"
	"github.com/tkeel-io/tkeel/readutil"
)

var (
	_log = logger.NewLogger("keel.service.keel")
)

type Keel struct {
	plugin *tkeel.Plugin
}

func New(p *tkeel.Plugin) (*Keel, error) {
	return &Keel{
		plugin: p,
	}, nil
}

func (k *Keel) Run() {
	pid := k.plugin.Conf().Plugin.ID
	if pid == "" {
		_log.Fatal("error plugin id: %s", pid)
	}
	if pid != "keel" {
		_log.Fatalf("error plugin id: %s should be keel", pid)
	}

	go func() {
		k.plugin.SetRequiredFunc(openapi.RequiredFunc{
			Identify: k.identify,
		})
		k.plugin.SetOptionalFunc(openapi.OptionalFunc{
			AddonsIdentify: k.addonsIdentify,
		})
		err := k.plugin.Run(&openapi.API{Endpoint: "/", H: k.Route})
		if err != nil {
			_log.Fatalf("error plugin run: %s", err)
			return
		}
	}()
	_log.Debug("keel running")
}

func (k *Keel) Route(e *openapi.APIEvent) {
	// check path.
	path := e.HTTPReq.RequestURI
	next, err := checkRoutePath(path)
	if err != nil {
		_log.Error(err)
		http.Error(e, "bad request", http.StatusBadRequest)
		return
	}
	if !next {
		e.WriteHeader(http.StatusOK)
	}

	pluginID, err := auth(e)
	if err != nil {
		_log.Errorf("error auth: %s", err)
		http.Error(e, err.Error(), http.StatusBadRequest)
		return
	}

	_log.Debugf("route plugin(%s) request %s", pluginID, e.HTTPReq.RequestURI)

	// find upstream plugin.
	upPluginID, endpoint, err := getUpstreamPlugin(e.HTTPReq.Context(), pluginID, path)
	if err != nil {
		_log.Errorf("error request(%s): %s", path, err)
		http.Error(e, err.Error(), http.StatusBadRequest)
	}

	// check upstream plugin.
	err = checkPluginStatus(e.HTTPReq.Context(), upPluginID)
	if err != nil {
		_log.Errorf("error check plugin(%s) status: %s", upPluginID, err)
		http.Error(e, "bad request", http.StatusBadRequest)
		return
	}

	resp, err := proxy(e.HTTPReq, e.ResponseWriter, upPluginID, endpoint)
	if err != nil {
		_log.Errorf("error proxy: %s", err)
		http.Error(e, err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()
	copyHeader(e.Header(), resp.Header)
	respBodyByte, err := io.ReadAll(resp.Body)
	if err != nil {
		_log.Errorf("error get response body: %s", err)
		http.Error(e, "bad request", http.StatusBadRequest)
		return
	}
	if _, err = e.Write(respBodyByte); err != nil {
		_log.Errorf("error response write: %s", err)
		return
	}
	_log.Debugf("route success.")
}

func (k *Keel) identify() (*openapi.IdentifyResp, error) {
	return &openapi.IdentifyResp{
		CommonResult: openapi.SuccessResult(),
		PluginID:     k.plugin.GetIdentifyResp().PluginID,
		Version:      k.plugin.GetIdentifyResp().Version,
		AddonsPoints: []*openapi.AddonsPoint{
			{
				AddonsPoint: "externalPreRouteCheck",
				Desc: `
				callback before external flow routing
				input request header and path
				output http statecode
				200   -- allow
				other -- deny
				`,
			},
		},
	}, nil
}

func (k *Keel) addonsIdentify(air *openapi.AddonsIdentifyReq) (*openapi.AddonsIdentifyResp, error) {
	endpointReq := air.Endpoint[0]
	xKeelStr := genBoolStr()

	resp, err := keel.CallKeel(context.TODO(), air.Plugin.ID, endpointReq.Endpoint,
		http.MethodGet, &keel.CallReq{
			Header: http.Header{
				"x-keel-check": []string{xKeelStr},
			},
			Body: []byte(`Check whether the endpoint correctly implements this callback:
			For example, when the request header contains the "x-keel-check" field, 
			the HTTP request header 200 is returned. When the field value is "True", 
			the body is 
			{	
				"msg":"ok",
				"ret":0
			}, 
			When it is False, the body is 
			{
				"msg":"faild",
				"ret":-1
			}. 
			If it is not included, it will judge whether the request is valid.`),
		})
	if err != nil {
		_log.Errorf("error addons identify: %w", err)
		return &openapi.AddonsIdentifyResp{
			CommonResult: openapi.BadRequestResult(resp.Status),
		}, nil
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			_log.Errorf("error response body close: %s", err)
		}
	}()
	result := &openapi.CommonResult{}
	if err := readutil.ReaderToJSON(resp.Body, result); err != nil {
		_log.Errorf("error read addons identify(%s/%s/%s) resp: %s",
			air.Plugin.ID, endpointReq.Endpoint, endpointReq.AddonsPoint, err.Error())
		return &openapi.AddonsIdentifyResp{
			CommonResult: openapi.BadRequestResult(err.Error()),
		}, nil
	}
	if (xKeelStr == "True" && result.Ret == 0 && result.Msg == "ok") ||
		(xKeelStr == "false" && result.Ret == -1 && result.Msg == "faild") {
		return &openapi.AddonsIdentifyResp{
			CommonResult: openapi.SuccessResult(),
		}, nil
	}
	_log.Errorf("error identify(%s/%s/%s) resp: %v",
		air.Plugin.ID, endpointReq.Endpoint, endpointReq.AddonsPoint, result)
	return &openapi.AddonsIdentifyResp{
		CommonResult: openapi.BadRequestResult(resp.Status),
	}, nil
}

func genBoolStr() string {
	n, err := rand.Int(rand.Reader, big.NewInt(100))
	if err != nil {
		_log.Errorf("error rand: %w", err)
		return "False"
	}
	if n.Int64()%2 == 1 {
		return "True"
	}
	return "False"
}

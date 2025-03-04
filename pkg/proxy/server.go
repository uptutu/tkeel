/*
Copyright 2021 The tKeel Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package proxy

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/tkeel-io/tkeel/pkg/model"
)

type Reqeust struct {
	ID         string      `json:"id"`
	Method     string      `json:"method"`
	Verb       string      `json:"verb"`
	Header     http.Header `json:"header"`
	QueryValue url.Values  `json:"query_value"`
	Body       []byte      `json:"body"`
}

func (p *Reqeust) String() string {
	b, err := json.Marshal(p)
	if err != nil {
		return err.Error()
	}
	return string(b)
}

type PluginProxyServer interface {
	// Watch watch for changes in the plugin proxy route map.
	// Call the function and pass in the changed plugin proxy route map when it changes.
	Watch(context.Context, func(model.PluginProxyRouteMap) error) error
	// ProxyPlugin proxy plugin request.
	ProxyPlugin(ctx context.Context, resp http.ResponseWriter, req *Reqeust) error
	// ProxyCore proxy core request.
	ProxyCore(ctx context.Context, resp http.ResponseWriter, req *http.Request) error
	// ProxyRudder proxy rudder request.
	ProxyRudder(ctx context.Context, resp http.ResponseWriter, req *http.Request) error
	// ProxySecurity proxy security request.
	ProxySecurity(ctx context.Context, resp http.ResponseWriter, req *http.Request) error
}

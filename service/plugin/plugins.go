package plugin

import (
	"context"
	"crypto/rand"
	"flag"
	"math/big"
	"net/http"
	"time"

	"github.com/tkeel-io/tkeel"
	"github.com/tkeel-io/tkeel/envutil"
	"github.com/tkeel-io/tkeel/keel"
	"github.com/tkeel-io/tkeel/logger"
	"github.com/tkeel-io/tkeel/openapi"
	"github.com/tkeel-io/tkeel/readutil"
	"github.com/tkeel-io/tkeel/token"
)

var (
	log = logger.NewLogger("keel.service.plugins")

	pluginsScrapeInterval = flag.String("plugins-scrape-interval", "30m",
		"The interval for the plugins to scrape the status of the registered plugin")
	pluginTokenSecret = flag.String("plugins-token-secret", envutil.Get("PLUGIN_TOKEN_SECRET", "changeme"), "gen token")

	idProvider token.IDProvider
)

type reqPlugin struct {
	*keel.Plugin      `json:",inline"`
	*keel.PluginRoute `json:",inline"`
}

type Manager struct {
	p *tkeel.Plugin
}

func New(p *tkeel.Plugin) (*Manager, error) {
	return &Manager{
		p: p,
	}, nil
}

func (m *Manager) Run() {
	pluginID := m.p.Conf().Plugin.ID
	if pluginID == "" {
		log.Fatalf("error plugin id: %s", pluginID)
	}
	if pluginID != "plugins" {
		log.Fatalf("error plugin id: %s should be plugins", pluginID)
	}

	idProvider = token.InitIDProvider([]byte(*pluginTokenSecret), "", "")

	go func() {
		scrapeInterval, err := time.ParseDuration(*pluginsScrapeInterval)
		if err != nil {
			log.Fatalf("error parse manager-scrape-interval: %s", err)
		}
		interval := scrapeInterval
		tick := time.NewTicker(interval)
		for range tick.C {
			scrapePluginStatus(context.TODO(), scrapeInterval)
			n, err := rand.Int(rand.Reader, big.NewInt(30))
			if err != nil {
				n = big.NewInt(30)
			}
			interval = time.Duration(n.Uint64())*time.Second + scrapeInterval
			tick.Reset(interval)
		}
	}()

	go func() {
		err := m.p.Run([]*openapi.API{
			{Endpoint: "/get", H: m.GetPlugins},
			{Endpoint: "/list", H: m.ListPlugins},
			{Endpoint: "/delete", H: m.DeletePlugins},
			{Endpoint: "/register", H: m.RegisterPlugins},
			{Endpoint: "/tenant-bind", H: m.TenantBind},
			{Endpoint: "/oauth2/token", H: m.Oauth2},
		}...)
		if err != nil {
			log.Fatalf("error plugin run: %s", err)
			return
		}
	}()

	log.Debugf("wait for dapr ready: %s", time.Now().Format(time.RFC3339Nano))
	if !keel.WaitDaprSidecarReady(10) {
		log.Fatalf("error dapr not ready")
	}
	log.Debugf("dapr ready: %s", time.Now().Format(time.RFC3339Nano))

	err := registerPlugin(context.TODO(), m.p.GetIdentifyResp(), *pluginTokenSecret)
	if err != nil {
		log.Debugf("error register plugin plugins: %s, If its a duplicate registration error, you can ignore it", err)
	}
	log.Debugf("register plugins ok")
}

func (m *Manager) GetPlugins(e *openapi.APIEvent) {
	log.Debugf("request plugins/get")

	if e.HTTPReq.Method != http.MethodGet {
		log.Errorf("error method(%s) not allow", e.HTTPReq.Method)
		http.Error(e, "method not allow", http.StatusMethodNotAllowed)
		return
	}

	pluginID := e.HTTPReq.URL.Query().Get("id")

	if pluginID == "" {
		http.Error(e, "plugin not registered", http.StatusBadRequest)
		return
	}

	getP, _, err := keel.GetPlugin(e.HTTPReq.Context(), pluginID)
	if err != nil {
		log.Errorf("error get plugins: %s", err)
		http.Error(e, "error get plugin", http.StatusInternalServerError)
		return
	}
	if getP == nil {
		http.Error(e, "plugin not registered", http.StatusBadRequest)
		return
	}

	pRoute, _, err := keel.GetPluginRoute(e.HTTPReq.Context(), pluginID)
	if err != nil {
		log.Errorf("error get plugins route: %s", err)
		http.Error(e, "error get plugin route", http.StatusInternalServerError)
		return
	}

	resp := &struct {
		openapi.CommonResult `json:",inline"`
		Data                 *reqPlugin `json:"data"`
	}{
		CommonResult: openapi.SuccessResult(),
		Data: &reqPlugin{
			Plugin:      getP,
			PluginRoute: pRoute,
		},
	}

	ret := e.ResponseJSON(resp)
	log.Debugf("get plugins: %s", string(ret))
}

func (m *Manager) ListPlugins(e *openapi.APIEvent) {
	log.Debugf("request plugins/list")

	if e.HTTPReq.Method != http.MethodGet {
		log.Errorf("error method(%s) not allow", e.HTTPReq.Method)
		http.Error(e, "method not allow", http.StatusMethodNotAllowed)
		return
	}

	allMap, _, err := keel.GetAllRegisteredPlugin(e.HTTPReq.Context())
	if err != nil {
		log.Errorf("error get all registered plugin: %s", err)
		http.Error(e, "error get all registered plugin", http.StatusInternalServerError)
		return
	}

	resp := &struct {
		openapi.CommonResult `json:",inline"`
		Data                 []*reqPlugin `json:"data"`
	}{
		CommonResult: openapi.SuccessResult(),
		Data:         make([]*reqPlugin, 0, len(allMap)),
	}

	for pID := range allMap {
		getP, _, err := keel.GetPlugin(e.HTTPReq.Context(), pID)
		if err != nil {
			log.Errorf("error get plugins: %s", err)
			continue
		}
		pRoute, _, err := keel.GetPluginRoute(e.HTTPReq.Context(), pID)
		if err != nil {
			log.Errorf("error get plugins route: %s", err)
			return
		}
		resp.Data = append(resp.Data, &reqPlugin{
			Plugin:      getP,
			PluginRoute: pRoute,
		})
	}

	ret := e.ResponseJSON(resp)
	log.Debugf("get plugins: %s", string(ret))
}

func (m *Manager) DeletePlugins(e *openapi.APIEvent) {
	log.Debugf("request plugins/delete")

	if e.HTTPReq.Method != http.MethodPost {
		log.Errorf("error method(%s) not allow", e.HTTPReq.Method)
		http.Error(e, "method not allow", http.StatusMethodNotAllowed)
		return
	}

	req := &struct {
		ID string `json:"id"`
	}{}
	err := readutil.ReaderToJSON(e.HTTPReq.Body, req)
	if err != nil {
		log.Errorf("error delete plugin: %s", err)
		http.Error(e, "error read request body", http.StatusBadRequest)
		return
	}
	defer e.HTTPReq.Body.Close()

	// delete plugin and all registered plugin.
	allPlugins, allEtag, err := keel.GetAllRegisteredPlugin(e.HTTPReq.Context())
	if err != nil {
		log.Errorf("error delete plugin: %s", err)
		http.Error(e, "error get all registered", http.StatusInternalServerError)
		return
	}
	pID := req.ID

	delP, _, err := keel.GetPlugin(e.HTTPReq.Context(), pID)
	if err != nil {
		log.Errorf("error get plugin(%s): %s", pID, err)
		http.Error(e, "error get plugin", http.StatusInternalServerError)
		return
	}
	if delP == nil {
		log.Debugf("delete plugin(%s) not registered", pID)
		e.ResponseJSON(openapi.SuccessResult())
		return
	}
	err = deleteMainPluginRoute(e.HTTPReq.Context(), pID, delP.MainPlugins)
	if err != nil {
		log.Errorf("error delete main plugin route: %w", err)
		http.Error(e, "internal error", http.StatusInternalServerError)
		return
	}
	err = keel.DeletePlugin(e.HTTPReq.Context(), pID)
	if err != nil {
		log.Errorf("error delete plugin(%s): %s", pID, err)
		http.Error(e, "error delete plugin", http.StatusInternalServerError)
		return
	}
	err = keel.DeletePluginRoute(e.HTTPReq.Context(), pID)
	if err != nil {
		log.Errorf("error delete plugin(%s) route: %s", pID, err)
		http.Error(e, "error delete plugin route", http.StatusInternalServerError)
		return
	}
	delete(allPlugins, pID)

	err = keel.SaveAllRegisteredPlugin(e.HTTPReq.Context(), allPlugins, allEtag)
	if err != nil {
		log.Errorf("error delete plugin: %s", err)
		http.Error(e, "error save all registered plugin", http.StatusInternalServerError)
		return
	}
	// return http succss.
	e.ResponseJSON(openapi.SuccessResult())
	log.Debugf("delete plugins: %v", req.ID)
}

func (m *Manager) RegisterPlugins(e *openapi.APIEvent) {
	log.Debugf("request plugins/register")

	if e.HTTPReq.Method != http.MethodPost {
		log.Errorf("error method(%s) not allow", e.HTTPReq.Method)
		http.Error(e, "method not allow", http.StatusMethodNotAllowed)
		return
	}

	req := &struct {
		ID     string `json:"id"`
		Secret string `json:"secret"`
	}{}
	err := readutil.ReaderToJSON(e.HTTPReq.Body, req)
	if err != nil {
		log.Errorf("error register plugin: %s", err)
		http.Error(e, "error read reqeuest body", http.StatusBadRequest)
		return
	}
	defer e.HTTPReq.Body.Close()

	pID := req.ID
	secret := req.Secret

	identifyResp, err := requestPluginIdentify(e.HTTPReq.Context(), pID)
	if err != nil {
		log.Errorf("error request plugins identify: %s", err)
		http.Error(e, "error requst new plugin", http.StatusBadRequest)
		return
	}

	err = registerPlugin(e.HTTPReq.Context(), identifyResp, secret)
	if err != nil {
		log.Errorf("error register plugins: %s", err)
		http.Error(e, "error register", http.StatusInternalServerError)
		return
	}
	// return http succss.
	e.ResponseJSON(openapi.SuccessResult())
	log.Debugf("register plugins: %s succ", pID)
}

func (m *Manager) TenantBind(e *openapi.APIEvent) {
	log.Debugf("request plugins/tenant-bind")

	if e.HTTPReq.Method != http.MethodPost {
		log.Errorf("error method(%s) not allow", e.HTTPReq.Method)
		http.Error(e, "method not allow", http.StatusMethodNotAllowed)
		return
	}

	req := &struct {
		PluginID string `json:"plugin_id"`
		Version  string `json:"version"`
		TenantID string `json:"tenant_id"`
		Extra    []byte `json:"extra"`
	}{}
	err := readutil.ReaderToJSON(e.HTTPReq.Body, req)
	if err != nil {
		log.Errorf("error register plugin: %s", err)
		http.Error(e, "error read request body", http.StatusBadRequest)
		return
	}
	defer e.HTTPReq.Body.Close()

	getP, getPluginEtag, err := keel.GetPlugin(e.HTTPReq.Context(), req.PluginID)
	if err != nil {
		log.Errorf("error get plugins: %s", err)
		http.Error(e, "error get plugins", http.StatusBadRequest)
		return
	}
	if getP == nil {
		http.Error(e, "error plugin not registered", http.StatusBadRequest)
		return
	}

	resp, err := requestPluginTenantBind(e.HTTPReq.Context(), req.PluginID, req.TenantID, req.Extra)
	if err != nil {
		log.Errorf("error request bind tenant: %s", err)
		http.Error(e, "error request bind tenant", http.StatusInternalServerError)
		return
	}
	if resp.Ret != 0 {
		log.Errorf("error request bind tenant: %s", resp.Msg)
		http.Error(e, "error reqeust plugins: "+resp.Msg, http.StatusBadRequest)
		return
	}

	if getP.ActiveTenant == nil {
		getP.ActiveTenant = make([]*keel.Tenant, 0, 1)
	}

	getP.ActiveTenant = append(getP.ActiveTenant, &keel.Tenant{
		TenantID: req.PluginID,
	})

	err = keel.SavePlugin(e.HTTPReq.Context(), getP, getPluginEtag)
	if err != nil {
		log.Errorf("error save plugin: %s", err)
		http.Error(e, "error save plugin", http.StatusInternalServerError)
		return
	}

	// return http succss.
	e.ResponseJSON(openapi.SuccessResult())
	log.Debugf("bind tenant plugins: %s/%s", req.PluginID, req.TenantID)
}

func (m *Manager) Oauth2(e *openapi.APIEvent) {
	log.Debug("==start== route oauth2")
	// get oauth2 request.
	pluginID, pluginSecret, err := parseOauth2Req(e.HTTPReq)
	if err != nil {
		log.Errorf("error parsh oauth2 request: %s", err)
		http.Error(e, "internal error", http.StatusInternalServerError)
		return
	}

	if pluginID == "" {
		log.Debug("plugins id is empty")
		http.Error(e, "bad request", http.StatusBadRequest)
		return
	}

	var nSecret string
	if pluginID == "plugins" {
		nSecret = *pluginTokenSecret
	} else {
		// get plugin state secret.
		plugin, _, err1 := keel.GetPlugin(e.HTTPReq.Context(), pluginID)
		if err1 != nil {
			log.Errorf("error get plugin: %s", err1)
			http.Error(e, "bad request", http.StatusBadRequest)
			return
		}
		if plugin == nil {
			log.Error("error plugin not registered")
			http.Error(e, "bad request", http.StatusBadRequest)
			return
		}
		nSecret = plugin.Secret
	}

	// check secret.
	if err = checkPluginSecret(nSecret, pluginSecret); err != nil {
		log.Errorf("error plugin secret: %s", err)
		http.Error(e, "bad request", http.StatusBadRequest)
		return
	}

	// gen access token.
	token, _, err := genPluginToken(pluginID)
	if err != nil {
		log.Errorf("error gen token: %s", err)
		http.Error(e, "internal error", http.StatusInternalServerError)
		return
	}

	e.ResponseJSON(struct {
		AccessToken  string `json:"access_token"`
		TokenType    string `json:"token_type"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int32  `json:"expires_in"`
	}{
		AccessToken: token,
		ExpiresIn:   int32((24 * time.Hour).Seconds()),
	})
	log.Debugf("issue plugin(%s/%s) token: %s", pluginID, pluginSecret, token)
}

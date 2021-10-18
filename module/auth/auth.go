package auth

import (
	"time"

	"github.com/tkeel-io/tkeel/logger"
	"github.com/tkeel-io/tkeel/module"
	"github.com/tkeel-io/tkeel/module/auth/api"
	"github.com/tkeel-io/tkeel/module/auth/model"
	"github.com/tkeel-io/tkeel/openapi"
)

var (
	log = logger.NewLogger("Keel.PluginAuth")
)

type PluginAuth struct {
	p   *module.Plugin
	api api.API
}

func NewPluginAuth(p *module.Plugin) *PluginAuth {
	authAPI := api.NewAPI()
	return &PluginAuth{p, authAPI}
}

func (p *PluginAuth) Run() {
	pluginID := p.p.Conf().Plugin.ID
	if pluginID == "" {
		log.Fatalf("error plugin id: %s", pluginID)
	}
	if pluginID != "auth" {
		log.Fatalf("error plugin id: %s should be auth", pluginID)
	}
	go func() {
		err := p.p.Run([]*openapi.API{
			{Endpoint: "/role/create", H: p.api.RoleCreate},
			{Endpoint: "/user/login", H: p.api.Login},
			{Endpoint: "/authenticate", H: p.api.Authenticate},
			{Endpoint: "/user/logout", H: p.api.UserLogout},
			{Endpoint: "/user/create", H: p.api.UserCreate},
			{Endpoint: "/tenant/create", H: p.api.TenantCreate},
			{Endpoint: "/tenant/list", H: p.api.TenantQuery},
			{Endpoint: "/token/parse", H: p.api.TokenParse},
			{Endpoint: "/token/create", H: p.api.TokenCreate},
			{Endpoint: "/token/valid", H: p.api.TokenValid},
		}...)
		if err != nil {
			log.Fatalf("error plugin run: %s", err)
			return
		}
	}()
	time.Sleep(2 * time.Minute)
	model.UserInit()
}

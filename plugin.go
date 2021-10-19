package tkeel

import (
	"flag"
	"fmt"
	"strconv"

	"github.com/tkeel-io/tkeel/envutil"
	"github.com/tkeel-io/tkeel/keel"
	"github.com/tkeel-io/tkeel/logger"
	"github.com/tkeel-io/tkeel/openapi"
	"github.com/tkeel-io/tkeel/version"
)

var (
	_log = logger.NewLogger("keel.plugin")

	_pluginID              = flag.String("plugin-id", envutil.Get("PLUGIN_ID", "keel-hello"), "Plugin id")
	_pluginVersion         = flag.String("plugin-version", envutil.Get("PLUGIN_VERSION", version.Version()), "Plugin version")
	_pluginHTTPPort        = flag.Int("plugin-http-port", _defaultPluginHTTPPort, "The port that the plugin listens to")
	_daprPort              = flag.String("dapr-http-port", envutil.Get("DAPR_HTTP_PORT", "3500"), "The port that the dapr listens to")
	_config                = flag.String("keel-plugin-_config", envutil.Get("KEEL_PLUGIN_CONFIG", ""), "Path to _config file, or name of a configuration object")
	_defaultPluginHTTPPort = 8080
)

type Plugin struct {
	conf *keel.Configuration
	*openapi.Openapi
}

func (p *Plugin) Conf() keel.Configuration {
	if p.conf != nil {
		return *p.conf
	}
	return keel.Configuration{}
}

func (p *Plugin) Run(apis ...*openapi.API) error {
	for _, a := range apis {
		p.AddOpenAPI(a)
	}
	if err := p.Listen(); err != nil {
		return fmt.Errorf("error plugin listen: %w", err)
	}
	return nil
}

func NewPluginFromFlags() (*Plugin, error) {
	if port, err := strconv.Atoi(envutil.Get("PLUGIN_HTTP_PORT", "8080")); err != nil {
		_log.Warn("using default HTTP Port 8080 for plugin")
	} else {
		_defaultPluginHTTPPort = port
	}

	flag.Parse()

	if !keel.K8SEnable {
		keel.SetDaprAddr("localhost:" + *_daprPort)
	}

	p, err := newPlugin()
	if err != nil {
		return nil, fmt.Errorf("new plugin err: %w", err)
	}
	return p, nil
}

func newPlugin() (*Plugin, error) {
	plugin := &Plugin{}

	if *_config != "" {
		conf, err := keel.LoadStandaloneConfiguration(*_config)
		if err != nil {
			_log.Errorf("load plugin _config(%s) err: %s", *_config, err)
			return nil, fmt.Errorf("error load plugin: %w", err)
		}
		plugin.conf = conf
		plugin.Openapi = openapi.NewOpenapi(plugin.conf.Plugin.Port, plugin.conf.Plugin.ID, plugin.conf.Plugin.Version)
		return plugin, nil
	}
	conf := keel.LoadDefaultConfiguration()
	conf.Plugin.ID = *_pluginID
	conf.Plugin.Port = *_pluginHTTPPort
	conf.Plugin.Version = *_pluginVersion
	plugin.conf = conf
	plugin.Openapi = openapi.NewOpenapi(plugin.conf.Plugin.Port, plugin.conf.Plugin.ID, plugin.conf.Plugin.Version)
	return plugin, nil
}

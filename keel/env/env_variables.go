package config

const (
	// PluginPort Port is the port of the plugin, grpc/http depending on mode.
	PluginPort string = "PLUGIN_PORT"
	// PluginID is the ID of the plugins.
	PluginID string = "PLUGIN_ID"
	// DaprGRPCPort is the dapr api grpc port.
	DaprGRPCPort string = "DAPR_GRPC_PORT"
	// DaprHTTPPort is the dapr api http port.
	DaprHTTPPort string = "DAPR_HTTP_PORT"
)

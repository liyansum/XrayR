package controller

import (
	"encoding/json"

	"github.com/xtls/xray-core/common/errors"

	"github.com/xtls/xray-core/core"
	"github.com/xtls/xray-core/infra/conf"

	"github.com/wyx2685/XrayR/api"
)

// OutboundBuilder build freedom outbound config for addOutbound
func OutboundBuilder(config *Config, nodeInfo *api.NodeInfo, tag string) (*core.OutboundHandlerConfig, error) {
	outboundDetourConfig := &conf.OutboundDetourConfig{}
	outboundDetourConfig.Protocol = "freedom"
	outboundDetourConfig.Tag = tag

	// SendThrough setting
	outboundDetourConfig.SendThrough = &config.SendIP

	// Freedom Protocol setting
	var domainStrategy = "Asis"
	if config.EnableDNS {
		if config.DNSType != "" {
			domainStrategy = config.DNSType
		} else {
			domainStrategy = "UseIP"
		}
	}
	proxySetting := &conf.FreedomConfig{
		DomainStrategy: domainStrategy,
	}
	var setting json.RawMessage
	setting, err := json.Marshal(proxySetting)
	if err != nil {
		return nil, errors.New("marshal proxy ", nodeInfo.NodeType, " config failed: ").Base(err)
	}
	outboundDetourConfig.Settings = &setting
	return outboundDetourConfig.Build()
}

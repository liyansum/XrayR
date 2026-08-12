// Package controller Package generate the InboundConfig used by add inbound
package controller

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sagernet/sing-shadowsocks/shadowaead_2022"
	C "github.com/sagernet/sing/common"
	"github.com/xtls/xray-core/common/net"
	"github.com/xtls/xray-core/core"
	"github.com/xtls/xray-core/infra/conf"

	"github.com/wyx2685/XrayR/api"
)

// InboundBuilder build Inbound config for different protocol
func InboundBuilder(config *Config, nodeInfo *api.NodeInfo, tag string) (*core.InboundHandlerConfig, error) {
	inboundDetourConfig := &conf.InboundDetourConfig{}
	// Build Listen IP address
	if config.ListenIP != "" {
		ipAddress := net.ParseAddress(config.ListenIP)
		inboundDetourConfig.ListenOn = &conf.Address{Address: ipAddress}
	}

	// Build Port
	portList := &conf.PortList{
		Range: []conf.PortRange{{From: nodeInfo.Port, To: nodeInfo.Port}},
	}
	inboundDetourConfig.PortList = portList
	// Build Tag
	inboundDetourConfig.Tag = tag
	// SniffingConfig
	sniffingConfig := &conf.SniffingConfig{
		Enabled:      true,
		DestOverride: conf.StringList{"http", "tls", "quic", "fakedns"},
	}
	if config.DisableSniffing {
		sniffingConfig.Enabled = false
	}
	inboundDetourConfig.SniffingConfig = sniffingConfig

	var (
		protocol      string
		streamSetting *conf.StreamConfig
		setting       json.RawMessage
	)

	var proxySetting any
	// Build Protocol and Protocol setting
	switch nodeInfo.NodeType {
	case "Trojan":
		protocol = "trojan"
		// Enable fallback
		if config.EnableFallback {
			fallbackConfigs, err := buildTrojanFallbacks(config.FallBackConfigs)
			if err == nil {
				proxySetting = &conf.TrojanServerConfig{
					Fallbacks: fallbackConfigs,
				}
			} else {
				return nil, err
			}
		} else {
			proxySetting = &conf.TrojanServerConfig{}
		}
	case "Shadowsocks":
		protocol = "shadowsocks"
		cipher := strings.ToLower(nodeInfo.CypherMethod)

		proxySetting = &conf.ShadowsocksServerConfig{
			Cipher:   cipher,
			Password: nodeInfo.ServerKey, // shadowsocks2022 shareKey
		}

		proxySetting, _ := proxySetting.(*conf.ShadowsocksServerConfig)
		// shadowsocks must have a random password
		// shadowsocks2022's password == user PSK, thus should a length of string >= 32 and base64 encoder
		b := make([]byte, 32)
		rand.Read(b)
		randPasswd := hex.EncodeToString(b)
		if C.Contains(shadowaead_2022.List, cipher) {
			proxySetting.Users = append(proxySetting.Users, &conf.ShadowsocksUserConfig{
				Password: base64.StdEncoding.EncodeToString(b),
			})
		} else {
			proxySetting.Password = randPasswd
		}

		proxySetting.NetworkList = &conf.NetworkList{"tcp", "udp"}

	default:
		return nil, fmt.Errorf("unsupported node type: %s; supported types: Trojan and Shadowsocks", nodeInfo.NodeType)
	}

	setting, err := json.Marshal(proxySetting)
	if err != nil {
		return nil, fmt.Errorf("marshal proxy %s config failed: %s", nodeInfo.NodeType, err)
	}
	inboundDetourConfig.Protocol = protocol
	inboundDetourConfig.Settings = &setting

	// Build streamSettings
	streamSetting = new(conf.StreamConfig)
	transportProtocol := conf.TransportProtocol(nodeInfo.TransportProtocol)
	networkType, err := transportProtocol.Build()
	if err != nil {
		return nil, fmt.Errorf("convert TransportProtocol failed: %s", err)
	}

	tcpSetting := &conf.TCPConfig{
		AcceptProxyProtocol: config.EnableProxyProtocol,
		HeaderConfig:        nodeInfo.Header,
	}
	streamSetting.TCPSettings = tcpSetting
	streamSetting.Network = &transportProtocol

	// Build TLS settings.
	if nodeInfo.EnableTLS && config.CertConfig != nil && config.CertConfig.CertMode != "none" {
		streamSetting.Security = "tls"
		certFile, keyFile, err := getCertFile(config.CertConfig)
		if err != nil {
			return nil, err
		}
		tlsSettings := &conf.TLSConfig{
			RejectUnknownSNI: config.CertConfig.RejectUnknownSni,
		}
		tlsSettings.Certs = append(tlsSettings.Certs, &conf.TLSCertConfig{CertFile: certFile, KeyFile: keyFile, OcspStapling: 3600})
		streamSetting.TLSSettings = tlsSettings
	}

	// Support ProxyProtocol for any transport protocol
	if networkType != "tcp" && config.EnableProxyProtocol {
		sockoptConfig := &conf.SocketConfig{
			AcceptProxyProtocol: config.EnableProxyProtocol,
		}
		streamSetting.SocketSettings = sockoptConfig
	}
	inboundDetourConfig.StreamSetting = streamSetting

	return inboundDetourConfig.Build()
}

func getCertFile(certConfig *CertConfig) (certFile string, keyFile string, err error) {
	if certConfig.CertMode != "file" {
		return "", "", fmt.Errorf("unsupported certmode %q: only file is supported", certConfig.CertMode)
	}
	if certConfig.CertFile == "" || certConfig.KeyFile == "" {
		return "", "", fmt.Errorf("cert file path and key file path are required")
	}
	return certConfig.CertFile, certConfig.KeyFile, nil
}

func buildTrojanFallbacks(fallbackConfigs []*FallBackConfig) ([]*conf.TrojanInboundFallback, error) {
	if fallbackConfigs == nil {
		return nil, fmt.Errorf("you must provide FallBackConfigs")
	}

	trojanFallBacks := make([]*conf.TrojanInboundFallback, len(fallbackConfigs))
	for i, c := range fallbackConfigs {

		if c.Dest == "" {
			return nil, fmt.Errorf("dest is required for fallback failed")
		}

		var dest json.RawMessage
		dest, err := json.Marshal(c.Dest)
		if err != nil {
			return nil, fmt.Errorf("marshal dest %s config failed: %s", dest, err)
		}
		trojanFallBacks[i] = &conf.TrojanInboundFallback{
			Name: c.SNI,
			Alpn: c.Alpn,
			Path: c.Path,
			Dest: dest,
			Xver: c.ProxyProtocolVer,
		}
	}
	return trojanFallBacks, nil
}

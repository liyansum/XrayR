package controller_test

import (
	"testing"

	"github.com/wyx2685/XrayR/api"
	. "github.com/wyx2685/XrayR/service/controller"
)

func TestBuildV2ray(t *testing.T) {
	nodeInfo := &api.NodeInfo{
		NodeType:          "V2ray",
		NodeID:            1,
		Port:              1145,
		SpeedLimit:        0,
		AlterID:           2,
		TransportProtocol: "ws",
		Host:              "test.test.tk",
		Path:              "v2ray",
		EnableTLS:         false,
	}
	config := &Config{}
	_, err := InboundBuilder(config, nodeInfo, "test_tag")
	if err != nil {
		t.Error(err)
	}
}

func TestBuildTrojan(t *testing.T) {
	nodeInfo := &api.NodeInfo{
		NodeType:          "Trojan",
		NodeID:            1,
		Port:              1145,
		SpeedLimit:        0,
		AlterID:           2,
		TransportProtocol: "tcp",
		Host:              "trojan.test.tk",
		Path:              "v2ray",
		EnableTLS:         false,
	}
	config := &Config{}
	_, err := InboundBuilder(config, nodeInfo, "test_tag")
	if err != nil {
		t.Error(err)
	}
}

func TestBuildSS(t *testing.T) {
	nodeInfo := &api.NodeInfo{
		NodeType:          "Shadowsocks",
		NodeID:            1,
		Port:              1145,
		SpeedLimit:        0,
		AlterID:           2,
		TransportProtocol: "tcp",
		Host:              "test.test.tk",
		Path:              "v2ray",
		EnableTLS:         false,
		CypherMethod:      "aes-256-gcm",
	}
	config := &Config{}
	_, err := InboundBuilder(config, nodeInfo, "test_tag")
	if err != nil {
		t.Error(err)
	}
}

func TestRejectAutomaticCertificateMode(t *testing.T) {
	nodeInfo := &api.NodeInfo{
		NodeType:          "Trojan",
		Port:              1145,
		TransportProtocol: "tcp",
		EnableTLS:         true,
	}
	config := &Config{
		CertConfig: &CertConfig{CertMode: "dns"},
	}

	if _, err := InboundBuilder(config, nodeInfo, "test_tag"); err == nil {
		t.Fatal("automatic certificate mode must be rejected")
	}
}

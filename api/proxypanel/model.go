package proxypanel

import "encoding/json"

type Response struct {
	Status  string          `json:"status"`
	Code    int             `json:"code"`
	Data    json.RawMessage `json:"data"`
	Message string          `json:"message"`
}

type ShadowsocksNodeInfo struct {
	ID          int    `json:"id"`
	SpeedLimit  uint64 `json:"speed_limit"`
	ClientLimit int    `json:"client_limit"`
	Method      string `json:"method"`
	Port        uint32 `json:"port"`
}

type TrojanNodeInfo struct {
	ID          int    `json:"id"`
	IsUDP       bool   `json:"is_udp"`
	SpeedLimit  uint64 `json:"speed_limit"`
	ClientLimit int    `json:"client_limit"`
	PushPort    int    `json:"push_port"`
	TrojanPort  uint32 `json:"trojan_port"`
}

// NodeStatus Node status report
type NodeStatus struct {
	CPU    string `json:"cpu"`
	Mem    string `json:"mem"`
	Net    string `json:"net"`
	Disk   string `json:"disk"`
	Uptime int    `json:"uptime"`
}

type NodeOnline struct {
	UID int    `json:"uid"`
	IP  string `json:"ip"`
}

type TrojanUser struct {
	UID        int    `json:"uid"`
	Password   string `json:"password"`
	SpeedLimit uint64 `json:"speed_limit"`
}

type SSUser struct {
	UID        int    `json:"uid"`
	Password   string `json:"passwd"`
	SpeedLimit uint64 `json:"speed_limit"`
}

type UserTraffic struct {
	UID      int   `json:"uid"`
	Upload   int64 `json:"upload"`
	Download int64 `json:"download"`
}

type NodeRule struct {
	Mode  string         `json:"mode"`
	Rules []NodeRuleItem `json:"rules"`
}

type NodeRuleItem struct {
	ID      int    `json:"id"`
	Type    string `json:"type"`
	Pattern string `json:"pattern"`
}

type IllegalReport struct {
	UID    int    `json:"uid"`
	RuleID int    `json:"rule_id"`
	Reason string `json:"reason"`
}

type Certificate struct {
	Key string `json:"key"`
	Pem string `json:"pem"`
}

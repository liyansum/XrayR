package gov2panel

import "encoding/json"

type serverConfig struct {
	shadowsocks
	Port   string           `json:"port"`
	Host   string           `json:"host"`
	Sni    string           `json:"sni"`
	Routes []route          `json:"routes"`
	Header *json.RawMessage `json:"header"`
}

type shadowsocks struct {
	Encryption   string `json:"encryption"`
	Obfs         string `json:"obfs"`
	ObfsSettings struct {
		Path string `json:"path"`
		Host string `json:"host"`
	} `json:"obfs_settings"`
	ServerKey string `json:"server_key"`
}

type route struct {
	Id          int      `json:"id"`
	Match       []string `json:"match"`
	Action      string   `json:"action"`
	ActionValue string   `json:"action_value"`
}

type user struct {
	Id         int    `json:"id"`
	Uuid       string `json:"uuid"`
	SpeedLimit int    `json:"speed_limit"`
}

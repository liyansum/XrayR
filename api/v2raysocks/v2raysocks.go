package v2raysocks

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/bitly/go-simplejson"
	"github.com/go-resty/resty/v2"
	"github.com/sagernet/sing-shadowsocks/shadowaead_2022"
	C "github.com/sagernet/sing/common"

	"github.com/wyx2685/XrayR/api"
)

// APIClient create an api client to the panel.
type APIClient struct {
	client        *resty.Client
	APIHost       string
	NodeID        int
	Key           string
	NodeType      string
	SpeedLimit    float64
	DeviceLimit   int
	LocalRuleList []api.DetectRule
	ConfigResp    *simplejson.Json
	access        sync.Mutex
	eTags         map[string]string
}

// New create an api instance
func New(apiConfig *api.Config) *APIClient {

	client := api.NewPanelHTTPClient(apiConfig)

	// Create Key for each requests
	client.SetQueryParams(map[string]string{
		"node_id": strconv.Itoa(apiConfig.NodeID),
		"token":   apiConfig.Key,
	})
	// Read local rule list
	localRuleList := api.ReadLocalRuleList(apiConfig.RuleListPath)
	apiClient := &APIClient{
		client:        client,
		NodeID:        apiConfig.NodeID,
		Key:           apiConfig.Key,
		APIHost:       apiConfig.APIHost,
		NodeType:      apiConfig.NodeType,
		SpeedLimit:    apiConfig.SpeedLimit,
		DeviceLimit:   apiConfig.DeviceLimit,
		LocalRuleList: localRuleList,
		eTags:         make(map[string]string),
	}
	return apiClient
}

// Describe return a description of the client
func (c *APIClient) Describe() api.ClientInfo {
	return api.ClientInfo{APIHost: c.APIHost, NodeID: c.NodeID, Key: c.Key, NodeType: c.NodeType}
}

// Debug set the client debug for client
func (c *APIClient) Debug() {
	api.EnableSafeDebug(c.client)
}

func (c *APIClient) assembleURL(path string) string {
	return c.APIHost + path
}

func (c *APIClient) parseResponse(res *resty.Response, path string, err error) (*simplejson.Json, error) {
	return api.ParseJSONResponse(res, c.assembleURL(path), err, 400)
}

// GetNodeInfo will pull NodeInfo Config from panel
func (c *APIClient) GetNodeInfo() (nodeInfo *api.NodeInfo, err error) {
	var nodeType string
	switch c.NodeType {
	case "Trojan", "Shadowsocks":
		nodeType = strings.ToLower(c.NodeType)
	default:
		return nil, fmt.Errorf("unsupported Node type: %s", c.NodeType)
	}
	res, err := c.client.R().
		SetHeader("If-None-Match", c.eTags["config"]).
		SetQueryParams(map[string]string{
			"act":      "config",
			"nodetype": nodeType,
		}).
		ForceContentType("application/json").
		Get(c.APIHost)

	// Etag identifier for a specific version of a resource. StatusCode = 304 means no changed
	if res.StatusCode() == 304 {
		return nil, errors.New(api.NodeNotModified)
	}
	// update etag
	if res.Header().Get("Etag") != "" && res.Header().Get("Etag") != c.eTags["config"] {
		c.eTags["config"] = res.Header().Get("Etag")
	}

	// Etag identifier for a specific version of a resource. StatusCode = 304 means no changed
	if res.StatusCode() == 304 {
		return nil, errors.New(api.NodeNotModified)
	}
	// update etag
	if res.Header().Get("Etag") != "" && res.Header().Get("Etag") != c.eTags["config"] {
		c.eTags["config"] = res.Header().Get("Etag")
	}

	response, err := c.parseResponse(res, "", err)
	c.access.Lock()
	defer c.access.Unlock()
	c.ConfigResp = response
	if err != nil {
		return nil, err
	}

	switch c.NodeType {
	case "Trojan":
		nodeInfo, err = c.ParseTrojanNodeResponse(response)
	case "Shadowsocks":
		nodeInfo, err = c.ParseSSNodeResponse(response)
	default:
		return nil, fmt.Errorf("unsupported Node type: %s", c.NodeType)
	}

	if err != nil {
		res, _ := response.MarshalJSON()
		return nil, fmt.Errorf("Parse node info failed: %s, \nError: %s", api.RedactText(string(res)), api.RedactText(err.Error()))
	}

	return nodeInfo, nil
}

// GetUserList will pull user form panel
func (c *APIClient) GetUserList() (UserList *[]api.UserInfo, err error) {
	var nodeType string
	switch c.NodeType {
	case "Trojan", "Shadowsocks":
		nodeType = strings.ToLower(c.NodeType)
	default:
		return nil, fmt.Errorf("unsupported Node type: %s", c.NodeType)
	}
	res, err := c.client.R().
		SetHeader("If-None-Match", c.eTags["user"]).
		SetQueryParams(map[string]string{
			"act":      "user",
			"nodetype": nodeType,
		}).
		ForceContentType("application/json").
		Get(c.APIHost)

	// Etag identifier for a specific version of a resource. StatusCode = 304 means no changed
	if res.StatusCode() == 304 {
		return nil, errors.New(api.UserNotModified)
	}
	// update etag
	if res.Header().Get("Etag") != "" && res.Header().Get("Etag") != c.eTags["user"] {
		c.eTags["user"] = res.Header().Get("Etag")
	}

	// Etag identifier for a specific version of a resource. StatusCode = 304 means no changed
	if res.StatusCode() == 304 {
		return nil, errors.New(api.UserNotModified)
	}
	// update etag
	if res.Header().Get("Etag") != "" && res.Header().Get("Etag") != c.eTags["user"] {
		c.eTags["user"] = res.Header().Get("Etag")
	}

	response, err := c.parseResponse(res, "", err)
	if err != nil {
		return nil, err
	}
	numOfUsers := len(response.Get("data").MustArray())
	userList := make([]api.UserInfo, numOfUsers)
	for i := 0; i < numOfUsers; i++ {
		user := api.UserInfo{}
		user.UID = response.Get("data").GetIndex(i).Get("id").MustInt()
		switch c.NodeType {
		case "Shadowsocks":
			user.Email = response.Get("data").GetIndex(i).Get("shadowsocks_user").Get("secret").MustString()
			user.Passwd = response.Get("data").GetIndex(i).Get("shadowsocks_user").Get("secret").MustString()
			user.Method = response.Get("data").GetIndex(i).Get("shadowsocks_user").Get("cipher").MustString()
			user.SpeedLimit = response.Get("data").GetIndex(i).Get("shadowsocks_user").Get("speed_limit").MustUint64() * 1000000 / 8
			user.DeviceLimit = response.Get("data").GetIndex(i).Get("shadowsocks_user").Get("device_limit").MustInt()
		case "Trojan":
			user.UUID = response.Get("data").GetIndex(i).Get("trojan_user").Get("password").MustString()
			user.Email = response.Get("data").GetIndex(i).Get("trojan_user").Get("password").MustString()
			user.SpeedLimit = response.Get("data").GetIndex(i).Get("trojan_user").Get("speed_limit").MustUint64() * 1000000 / 8
			user.DeviceLimit = response.Get("data").GetIndex(i).Get("trojan_user").Get("device_limit").MustInt()
		}
		if c.SpeedLimit > 0 {
			user.SpeedLimit = uint64((c.SpeedLimit * 1000000) / 8)
		}

		if c.DeviceLimit > 0 {
			user.DeviceLimit = c.DeviceLimit
		}

		userList[i] = user
	}
	return &userList, nil
}

// ReportUserTraffic reports the user traffic
func (c *APIClient) ReportUserTraffic(userTraffic *[]api.UserTraffic) error {

	data := make([]UserTraffic, len(*userTraffic))
	for i, traffic := range *userTraffic {
		data[i] = UserTraffic{
			UID:      traffic.UID,
			Upload:   traffic.Upload,
			Download: traffic.Download}
	}

	res, err := c.client.R().
		SetQueryParam("node_id", strconv.Itoa(c.NodeID)).
		SetQueryParams(map[string]string{
			"act":      "submit",
			"nodetype": strings.ToLower(c.NodeType),
		}).
		SetBody(data).
		ForceContentType("application/json").
		Post(c.APIHost)
	_, err = c.parseResponse(res, "", err)
	if err != nil {
		return err
	}
	return nil
}

// GetNodeRule implements the API interface
func (c *APIClient) GetNodeRule() (*[]api.DetectRule, error) {
	ruleList := c.LocalRuleList

	// fix: reuse config response
	c.access.Lock()
	defer c.access.Unlock()
	ruleListResponse := c.ConfigResp.Get("routing").Get("rules").GetIndex(1).Get("domain").MustStringArray()
	for i, rule := range ruleListResponse {
		rule = strings.TrimPrefix(rule, "regexp:")
		ruleListItem := api.DetectRule{
			ID:      i,
			Pattern: regexp.MustCompile(rule),
		}
		ruleList = append(ruleList, ruleListItem)
	}
	return &ruleList, nil
}

// ReportNodeStatus implements the API interface
func (c *APIClient) ReportNodeStatus(nodeStatus *api.NodeStatus) (err error) {
	systemload := NodeStatus{
		Uptime: int(nodeStatus.Uptime),
		CPU:    fmt.Sprintf("%d%%", int(nodeStatus.CPU)),
		Mem:    fmt.Sprintf("%d%%", int(nodeStatus.Mem)),
		Disk:   fmt.Sprintf("%d%%", int(nodeStatus.Disk)),
	}

	res, err := c.client.R().
		SetQueryParam("node_id", strconv.Itoa(c.NodeID)).
		SetQueryParams(map[string]string{
			"act":      "nodestatus",
			"nodetype": strings.ToLower(c.NodeType),
		}).
		SetBody(systemload).
		ForceContentType("application/json").
		Post(c.APIHost)
	_, err = c.parseResponse(res, "", err)
	if err != nil {
		return err
	}
	return nil
}

// ReportNodeOnlineUsers implements the API interface
func (c *APIClient) ReportNodeOnlineUsers(onlineUserList *[]api.OnlineUser) error {
	data := make([]NodeOnline, len(*onlineUserList))
	for i, user := range *onlineUserList {
		data[i] = NodeOnline{UID: user.UID, IP: user.IP}
	}

	res, err := c.client.R().
		SetQueryParam("node_id", strconv.Itoa(c.NodeID)).
		SetQueryParams(map[string]string{
			"act":      "onlineusers",
			"nodetype": strings.ToLower(c.NodeType),
		}).
		SetBody(data).
		ForceContentType("application/json").
		Post(c.APIHost)
	_, err = c.parseResponse(res, "", err)
	if err != nil {
		return err
	}
	return nil
}

// ReportIllegal implements the API interface
func (c *APIClient) ReportIllegal(detectResultList *[]api.DetectResult) error {
	return nil
}

// ParseTrojanNodeResponse parse the response for the given nodeInfo format
func (c *APIClient) ParseTrojanNodeResponse(nodeInfoResponse *simplejson.Json) (*api.NodeInfo, error) {
	tmpInboundInfo := nodeInfoResponse.Get("inbounds").MustArray()
	marshalByte, _ := json.Marshal(tmpInboundInfo[0].(map[string]interface{}))
	inboundInfo, _ := simplejson.NewJson(marshalByte)

	port := uint32(inboundInfo.Get("port").MustUint64())
	host := inboundInfo.Get("streamSettings").Get("tlsSettings").Get("serverName").MustString()

	// Create GeneralNodeInfo
	nodeInfo := &api.NodeInfo{
		NodeType:          c.NodeType,
		NodeID:            c.NodeID,
		Port:              port,
		TransportProtocol: "tcp",
		EnableTLS:         true,
		Host:              host,
	}
	return nodeInfo, nil
}

// ParseSSNodeResponse parse the response for the given nodeInfo format
func (c *APIClient) ParseSSNodeResponse(nodeInfoResponse *simplejson.Json) (*api.NodeInfo, error) {
	var method, serverPsk string
	tmpInboundInfo := nodeInfoResponse.Get("inbounds").MustArray()
	marshalByte, _ := json.Marshal(tmpInboundInfo[0].(map[string]interface{}))
	inboundInfo, _ := simplejson.NewJson(marshalByte)

	port := uint32(inboundInfo.Get("port").MustUint64())
	method = inboundInfo.Get("settings").Get("method").MustString()
	// Shadowsocks 2022
	if C.Contains(shadowaead_2022.List, method) {
		serverPsk = inboundInfo.Get("settings").Get("password").MustString()
	}

	// Create GeneralNodeInfo
	nodeInfo := &api.NodeInfo{
		NodeType:          c.NodeType,
		NodeID:            c.NodeID,
		Port:              port,
		TransportProtocol: "tcp",
		CypherMethod:      method,
		ServerKey:         serverPsk,
	}

	return nodeInfo, nil
}

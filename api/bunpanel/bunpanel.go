package bunpanel

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/wyx2685/XrayR/api"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

type APIClient struct {
	client           *resty.Client
	APIHost          string
	NodeID           int
	Key              string
	NodeType         string
	SpeedLimit       float64
	DeviceLimit      int
	LocalRuleList    []api.DetectRule
	LastReportOnline map[int]int
	access           sync.Mutex
	eTags            map[string]string
}

// ReportIllegal implements api.API.
func (*APIClient) ReportIllegal(detectResultList *[]api.DetectResult) (err error) {
	return nil
}

// ReportNodeStatus implements api.API.
func (*APIClient) ReportNodeStatus(nodeStatus *api.NodeStatus) (err error) {
	return nil
}

// GetNodeRule implements api.API.
func (c *APIClient) GetNodeRule() (*[]api.DetectRule, error) {
	ruleList := c.LocalRuleList
	return &ruleList, nil
}

func New(apiConfig *api.Config) *APIClient {
	client := api.NewPanelHTTPClient(apiConfig)
	client.SetBaseURL(apiConfig.APIHost)
	// Create Key for each requests
	client.SetQueryParams(map[string]string{
		"serverId": strconv.Itoa(apiConfig.NodeID),
		"nodeType": strings.ToLower(apiConfig.NodeType),
		"token":    apiConfig.Key,
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

func (c *APIClient) parseResponse(res *resty.Response, path string, err error) (*Response, error) {
	if err := api.CheckResponse(res, c.assembleURL(path), err, 400); err != nil {
		return nil, err
	}
	response := res.Result().(*Response)

	if response.StatusCode != 200 {
		res, _ := json.Marshal(&response)
		return nil, fmt.Errorf("statusCode %s invalid", api.RedactText(string(res)))
	}
	return response, nil
}

func (c *APIClient) GetNodeInfo() (nodeInfo *api.NodeInfo, err error) {
	path := fmt.Sprintf("/v2/server/%d/get", c.NodeID)
	res, err := c.client.R().
		SetResult(&Response{}).
		SetHeader("If-None-Match", c.eTags["node"]).
		ForceContentType("application/json").
		Get(path)
	// Etag identifier for a specific version of a resource. StatusCode = 304 means no changed
	if res.StatusCode() == 304 {
		return nil, errors.New(api.NodeNotModified)
	}

	if res.Header().Get("ETag") != "" && res.Header().Get("ETag") != c.eTags["node"] {
		c.eTags["node"] = res.Header().Get("ETag")
	}

	response, err := c.parseResponse(res, path, err)
	if err != nil {
		return nil, err
	}

	nodeInfoResponse := new(Server)

	if err := json.Unmarshal(response.Datas, nodeInfoResponse); err != nil {
		return nil, fmt.Errorf("unmarshal %s failed: %s", reflect.TypeOf(nodeInfoResponse), err)
	}

	nodeInfo, err = c.ParseNodeInfo(nodeInfoResponse)
	if err != nil {
		res, _ := json.Marshal(nodeInfoResponse)
		return nil, fmt.Errorf("parse node info failed: %s, \nError: %s, \nPlease check the doc of custom_config for help: https://xrayr-project.github.io/XrayR-doc/dui-jie-sspanel/sspanel/sspanel_custom_config", api.RedactText(string(res)), api.RedactText(err.Error()))
	}

	if err != nil {
		res, _ := json.Marshal(nodeInfoResponse)
		return nil, fmt.Errorf("parse node info failed: %s, \nError: %s", api.RedactText(string(res)), api.RedactText(err.Error()))
	}

	return nodeInfo, nil
}

func (c *APIClient) GetUserList() (UserList *[]api.UserInfo, err error) {
	path := "/v2/user/get"
	res, err := c.client.R().
		SetQueryParam("serverId", strconv.Itoa(c.NodeID)).
		SetHeader("If-None-Match", c.eTags["users"]).
		SetResult(&Response{}).
		ForceContentType("application/json").
		Get(path)
	// Etag identifier for a specific version of a resource. StatusCode = 304 means no changed
	if res.StatusCode() == 304 {
		return nil, errors.New(api.UserNotModified)
	}

	if res.Header().Get("ETag") != "" && res.Header().Get("ETag") != c.eTags["users"] {
		c.eTags["users"] = res.Header().Get("ETag")
	}

	response, err := c.parseResponse(res, path, err)
	if err != nil {
		return nil, err
	}

	userListResponse := new([]User)

	if err := json.Unmarshal(response.Datas, userListResponse); err != nil {
		return nil, fmt.Errorf("unmarshal %s failed: %s", reflect.TypeOf(userListResponse), err)
	}
	userList, err := c.ParseUserListResponse(userListResponse)
	if err != nil {
		res, _ := json.Marshal(userListResponse)
		return nil, fmt.Errorf("parse user list failed: %s", api.RedactText(string(res)))
	}
	return userList, nil
}

func (c *APIClient) ReportNodeOnlineUsers(onlineUserList *[]api.OnlineUser) error {
	c.access.Lock()
	defer c.access.Unlock()

	reportOnline := make(map[int]int)
	data := make([]OnlineUser, len(*onlineUserList))
	for i, user := range *onlineUserList {
		data[i] = OnlineUser{UID: user.UID, IP: user.IP}
		reportOnline[user.UID]++
	}
	c.LastReportOnline = reportOnline // Update LastReportOnline

	postData := &PostData{Data: data}
	path := "/v2/user/online/create"
	res, err := c.client.R().
		SetQueryParam("serverId", strconv.Itoa(c.NodeID)).
		SetBody(postData).
		SetResult(&Response{}).
		ForceContentType("application/json").
		Post(path)

	_, err = c.parseResponse(res, path, err)
	if err != nil {
		return err
	}

	return nil
}

func (c *APIClient) ReportUserTraffic(userTraffic *[]api.UserTraffic) error {

	data := make([]UserTraffic, len(*userTraffic))
	for i, traffic := range *userTraffic {
		data[i] = UserTraffic{
			UID:      traffic.UID,
			Upload:   traffic.Upload,
			Download: traffic.Download}
	}
	postData := &PostData{Data: data}
	path := "/v2/user/data-usage/create"
	res, err := c.client.R().
		SetQueryParam("serverId", strconv.Itoa(c.NodeID)).
		SetBody(postData).
		SetResult(&Response{}).
		ForceContentType("application/json").
		Post(path)
	_, err = c.parseResponse(res, path, err)
	if err != nil {
		return err
	}

	return nil
}

func (c *APIClient) ParseUserListResponse(userInfoResponse *[]User) (*[]api.UserInfo, error) {
	c.access.Lock()
	// Clear Last report log
	defer func() {
		c.LastReportOnline = make(map[int]int)
		c.access.Unlock()
	}()

	var deviceLimit, localDeviceLimit = 0, 0
	var speedLimit uint64 = 0
	var userList []api.UserInfo
	for _, user := range *userInfoResponse {
		if c.DeviceLimit > 0 {
			deviceLimit = c.DeviceLimit
		} else {
			deviceLimit = user.DeviceLimit
		}

		// If there is still device available, add the user
		if deviceLimit > 0 && user.AliveIP > 0 {
			lastOnline := 0
			if v, ok := c.LastReportOnline[user.ID]; ok {
				lastOnline = v
			}
			// If there are any available device.
			if localDeviceLimit = deviceLimit - user.AliveIP + lastOnline; localDeviceLimit > 0 {
				deviceLimit = localDeviceLimit
				// If this backend server has reported any user in the last reporting period.
			} else if lastOnline > 0 {
				deviceLimit = lastOnline
				// Remove this user.
			} else {
				continue
			}
		}

		if c.SpeedLimit > 0 {
			speedLimit = uint64((c.SpeedLimit * 1000000) / 8)
		} else {
			speedLimit = uint64((user.SpeedLimit * 1000000) / 8)
		}
		userList = append(userList, api.UserInfo{
			UID:         user.ID,
			UUID:        user.UUID,
			SpeedLimit:  speedLimit,
			DeviceLimit: deviceLimit,
			Passwd:      user.UUID,
			Email:       user.UUID + "@bunpanel.user",
		})
	}

	return &userList, nil
}

func (c *APIClient) ParseNodeInfo(nodeInfoResponse *Server) (*api.NodeInfo, error) {
	nodeConfig := nodeInfoResponse
	var enableTLS bool
	switch c.NodeType {
	case "Shadowsocks":
	case "Trojan":
		enableTLS = true
	default:
		return nil, fmt.Errorf("unsupported node type: %s", c.NodeType)
	}

	return &api.NodeInfo{
		NodeType:          c.NodeType,
		NodeID:            c.NodeID,
		Port:              uint32(nodeConfig.Port),
		TransportProtocol: "tcp",
		EnableTLS:         enableTLS,
		CypherMethod:      nodeConfig.Method,
	}, nil
}

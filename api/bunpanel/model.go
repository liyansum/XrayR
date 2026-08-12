package bunpanel

import "encoding/json"

type Server struct {
	Port   int    `json:"serverPort"`
	Method string `json:"method"`
}

type User struct {
	ID          int     `json:"id"`
	UUID        string  `json:"uuid"`
	SpeedLimit  float64 `json:"speedLimit"`
	DeviceLimit int     `json:"ipLimit"`
	AliveIP     int     `json:"onlineIp"`
}

type OnlineUser struct {
	UID int    `json:"userId"`
	IP  string `json:"ip"`
}

// UserTraffic is the data structure of traffic
type UserTraffic struct {
	UID      int   `json:"userId"`
	Upload   int64 `json:"u"`
	Download int64 `json:"d"`
}

type Response struct {
	StatusCode int             `json:"statusCode"`
	Datas      json.RawMessage `json:"datas"`
}

type PostData struct {
	Data interface{} `json:"data"`
}

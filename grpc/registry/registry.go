package registry

type ServiceRegistry struct {
	AgentID         string   `json:"agentid"`
	PrivateIpv4     []string `json:"privateipv4"`
	PrivateIpv6     []string `json:"privateipv6"`
	Platform        string   `json:"platform"`
	PlatformFamily  string   `json:"platformfamily"`
	PlatformVersion string   `json:"platformversion"`
	KernelVersion   string   `json:"kernelversion"`
	CreateAt        int      `json:"createat"`
	EndAt           int      `json:"endat"`
	Version         string   `json:"version"`
	Online          bool     `json:"online"`
}

func (s *ServiceRegistry) Regist() {
	// if jsonBytes, err := json.Marshal(s); err == nil {
	// 	string(jsonBytes)
	// }
}

func (s *ServiceRegistry) Quit() {

}

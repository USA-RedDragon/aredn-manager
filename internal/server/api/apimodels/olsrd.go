package apimodels

type OlsrdLinks struct {
	PID                   int             `json:"pid"`
	SystemTime            uint64          `json:"systemTime"`
	TimeSinceStartup      uint64          `json:"timeSinceStartup"`
	ConfigurationChecksum string          `json:"configurationChecksum"`
	Links                 []OlsrdLinkinfo `json:"links"`
}

type OlsrdLinkinfo struct {
	HelloTime           uint64  `json:"helloTime"`
	LostLinkTime        uint64  `json:"lostLinkTime"`
	LinkQuality         float32 `json:"linkQuality"`
	VTime               uint64  `json:"vtime"`
	LinkCost            float32 `json:"linkCost"`
	LinkType            string  `json:"linkType"`
	Hostname            string  `json:"hostname"`
	PreviousLinkStatus  string  `json:"previousLinkStatus"`
	CurrentLinkStatus   string  `json:"currentLinkStatus"`
	NeighborLinkQuality uint64  `json:"neighborLinkQuality"`
	SymmetryTime        uint64  `json:"symmetryTime"`
	SeqnoValid          bool    `json:"seqnoValid"`
	Pending             bool    `json:"pending"`
	LossHelloInterval   uint64  `json:"lossHelloInterval"`
	LossMultiplier      uint64  `json:"lossMultiplier"`
	Hysteresis          float32 `json:"hysteresis"`
	Seqno               uint64  `json:"seqno"`
	LossTime            uint64  `json:"lossTime"`
	ValidityTime        uint64  `json:"validityTime"`
	OLSRInterface       string  `json:"olsrInterface"`
	LastHelloTime       uint64  `json:"lastHelloTime"`
	AsymmetryTime       uint64  `json:"asymmetryTime"`
	LocalIP             string  `json:"localIP"`
	RemoteIP            string  `json:"remoteIP"`
	InterfaceName       string  `json:"ifName"`
}

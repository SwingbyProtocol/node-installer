package bot

type BlockBook struct {
	BlockBook Response `json:"blockbook"`
	Backend   Backend  `json:"backend"`
}

type Response struct {
	Syncmode      bool `json:"syncMode"`
	InSync        bool `json:"inSync"`
	BestHeight    int  `json:"bestHeight"`
	MempoolSize   int  `json:"mempoolSize"`
	InSyncMempool bool `json:"inSyncMempool"`
}

type Backend struct {
	Blocks     int    `json:"blocks"`
	Headers    int    `json:"headers"`
	SizeOnDisk int    `json:"sizeOnDisk"`
	Version    string `json:"version"`
}

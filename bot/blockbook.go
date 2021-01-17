package bot

type BlockBook struct {
	BlockBook *Response `json:"blockbook"`
	Backend   *Backend  `json:"backend"`
}

type Response struct {
	Syncmode   bool `json:"syncMode"`
	InSync     bool `json:"inSync"`
	BestHeight int
	SFbSize    int
}

type Backend struct {
	Blocks     int
	Beaders    int
	SizeOnDisk int
}

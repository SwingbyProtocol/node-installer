package bot

type BlockBook struct {
	BlockBook *Response
	Backend   *Backend
}

type Response struct {
	Syncmode   bool
	InSync     bool
	BestHeight int
	SFbSize    int
}

type Backend struct {
	Blocks     int
	Beaders    int
	SizeOnDisk int
}

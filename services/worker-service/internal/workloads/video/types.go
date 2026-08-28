package video

type VideoParameters struct {
	Resolution        string `json:"resolution"`
	Codec             string `json:"codec"`
	Bitrate           string `json:"bitrate"`
	GenerateThumbnail bool   `json:"generateThumbnail"`
}

type VideoMetaData struct {
	DurationSeconds float64 `json:"durationSeconds"`
	Width           int     `json:"width"`
	Height          int     `json:"height"`
	CodecName       string  `json:"codecName"`
	Bitrate         int64   `json:"bitrate"`
}

type Season struct {
	Number       int       `json:"number_seasons"`
	EpisodeCount int       `json:"episode_count"`
	Episodes     []Episode `json:"episodes"`
}
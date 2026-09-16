type Anime struct {
	IDAnime           int      `json:"id_anime"`
	MainTitle         string   `json:"main_title"`
	NativeTitle       string   `json:"native_title"`
	AlternativeTitles []string `json:"alternative_titles"`
	Synopsis          string   `json:"synopsis"`
	NumberEpisodes    int      `json:"number_episodes"`
	EpisodeCount      int      `json:"episode_count"`
}
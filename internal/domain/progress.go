type Progress struct {
	AnimeID     int  `json:"anime_id"`
	EpisodeID   int  `json:"episode_id"`
	TimeWatched int  `json:"time_watched"`
	Completed   bool `json:"episode_completed"`
}
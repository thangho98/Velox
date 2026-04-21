package ophim

// PageResponse is the response from /v1/api/danh-sach/phim-moi-cap-nhat
type PageResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Data    struct {
		Items  []MovieItem `json:"items"`
		Params struct {
			Pagination struct {
				TotalItems        int `json:"totalItems"`
				TotalItemsPerPage int `json:"totalItemsPerPage"`
				CurrentPage       int `json:"currentPage"`
				TotalPages        int `json:"totalPages"`
			} `json:"pagination"`
		} `json:"params"`
	} `json:"data"`
}

// MovieItem represents a movie in the list
type MovieItem struct {
	ID             string `json:"_id"`
	Name           string `json:"name"`
	Slug           string `json:"slug"`
	OriginName     string `json:"origin_name"`
	Type           string `json:"type"`
	ThumbURL       string `json:"thumb_url"`
	PosterURL      string `json:"poster_url"`
	Year           int    `json:"year"`
	Time           string `json:"time"`
	Quality        string `json:"quality"`
	Lang           string `json:"lang"`
	EpisodeCurrent string `json:"episode_current"`
}

// MovieDetailResponse is the response from /phim/{slug}
type MovieDetailResponse struct {
	Status   bool      `json:"status"`
	Message  string    `json:"msg"`
	Movie    Movie     `json:"movie"`
	Episodes []Episode `json:"episodes"`
}

// Movie represents the detailed movie object
type Movie struct {
	ID             string   `json:"_id"`
	Name           string   `json:"name"`
	Slug           string   `json:"slug"`
	OriginName     string   `json:"origin_name"`
	Content        string   `json:"content"`
	Type           string   `json:"type"`
	Status         string   `json:"status"`
	ThumbURL       string   `json:"thumb_url"`
	PosterURL      string   `json:"poster_url"`
	Time           string   `json:"time"`
	EpisodeCurrent string   `json:"episode_current"`
	EpisodeTotal   string   `json:"episode_total"`
	Quality        string   `json:"quality"`
	Lang           string   `json:"lang"`
	Year           int      `json:"year"`
	Actor          []string `json:"actor"`
	Director       []string `json:"director"`
	Category       []struct {
		Name string `json:"name"`
	} `json:"category"`
	Country []struct {
		Name string `json:"name"`
	} `json:"country"`
}

// Episode represents the episode list for a server
type Episode struct {
	ServerName string       `json:"server_name"`
	ServerData []ServerData `json:"server_data"`
}

// ServerData contains the stream link for a specific episode
type ServerData struct {
	Name      string `json:"name"`
	Slug      string `json:"slug"`
	Filename  string `json:"filename"`
	LinkEmbed string `json:"link_embed"`
	LinkM3U8  string `json:"link_m3u8"`
}

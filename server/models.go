package server

type TraktTokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
	CreatedAt    int    `json:"created_at"`
}

type TraktWatchingResponse struct {
	Type      string   `json:"type"`
	Action    string   `json:"action"`
	Movie     *Movie   `json:"movie,omitempty"`
	Show      *Show    `json:"show,omitempty"`
	Episode   *Episode `json:"episode,omitempty"`
	StartedAt string   `json:"started_at"`
	ExpiresAt string   `json:"expires_at"`
}

type Movie struct {
	Title string `json:"title"`
	Year  int    `json:"year"`
	IDs   IDs    `json:"ids"`
}

type Show struct {
	Title string `json:"title"`
	Year  int    `json:"year"`
	IDs   IDs    `json:"ids"`
}

type Episode struct {
	Season int    `json:"season"`
	Number int    `json:"number"`
	Title  string `json:"title"`
	IDs    IDs    `json:"ids"`
}

type IDs struct {
	Trakt int    `json:"trakt"`
	TVDB  int    `json:"tvdb"`
	IMDB  string `json:"imdb"`
	TMDB  int    `json:"tmdb"`
	Slug  string `json:"slug,omitempty"`
}

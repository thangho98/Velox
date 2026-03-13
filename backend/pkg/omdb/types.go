package omdb

import "strconv"

// Response is the OMDb API response for a single title lookup.
type Response struct {
	Title    string   `json:"Title"`
	Year     string   `json:"Year"`
	Rated    string   `json:"Rated"`
	Released string   `json:"Released"`
	Runtime  string   `json:"Runtime"`
	Genre    string   `json:"Genre"`
	Director string   `json:"Director"`
	Writer   string   `json:"Writer"`
	Actors   string   `json:"Actors"`
	Plot     string   `json:"Plot"`
	Language string   `json:"Language"`
	Country  string   `json:"Country"`
	Awards   string   `json:"Awards"`
	Poster   string   `json:"Poster"`
	Ratings  []Rating `json:"Ratings"`
	// Direct fields
	Metascore  string `json:"Metascore"`
	IMDbRating string `json:"imdbRating"`
	IMDbVotes  string `json:"imdbVotes"`
	IMDbID     string `json:"imdbID"`
	Type       string `json:"Type"` // movie, series, episode
	BoxOffice  string `json:"BoxOffice"`

	// API status
	Response string `json:"Response"` // "True" or "False"
	Error    string `json:"Error"`
}

// Rating is a single rating source entry.
type Rating struct {
	Source string `json:"Source"`
	Value  string `json:"Value"`
}

// IMDbRatingFloat parses IMDb rating as float64 (0-10). Returns 0 on error.
func (r *Response) IMDbRatingFloat() float64 {
	f, err := strconv.ParseFloat(r.IMDbRating, 64)
	if err != nil {
		return 0
	}
	return f
}

// RottenTomatoesScore extracts the RT percentage (0-100). Returns 0 if not found.
func (r *Response) RottenTomatoesScore() int {
	for _, rating := range r.Ratings {
		if rating.Source == "Rotten Tomatoes" {
			// Value format: "93%"
			s := rating.Value
			if len(s) > 0 && s[len(s)-1] == '%' {
				n, err := strconv.Atoi(s[:len(s)-1])
				if err == nil {
					return n
				}
			}
		}
	}
	return 0
}

// MetascoreInt parses the Metacritic score (0-100). Returns 0 on error.
func (r *Response) MetascoreInt() int {
	n, err := strconv.Atoi(r.Metascore)
	if err != nil {
		return 0
	}
	return n
}

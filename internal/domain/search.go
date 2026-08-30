package domain

type SearchResult struct {
	Query   string    `json:"query"`
	Artists []*Artist `json:"artists"`
	Albums  []*Album  `json:"albums"`
	Tracks  []*Track  `json:"tracks"`
}

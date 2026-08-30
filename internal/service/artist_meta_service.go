package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"ne/internal/repository"
)

type ArtistMetaService struct {
	catalogRepo *repository.CatalogRepository
	httpClient  *http.Client
}

func NewArtistMetaService(catalogRepo *repository.CatalogRepository) *ArtistMetaService {
	return &ArtistMetaService{
		catalogRepo: catalogRepo,
		httpClient:  &http.Client{Timeout: 5 * time.Second},
	}
}

type ArtistBiography struct {
	ArtistID    string `json:"artistId"`
	Name        string `json:"name"`
	Biography   string `json:"biography"`
	ImageURL    string `json:"imageUrl,omitempty"`
	Source      string `json:"source"`
}

func (s *ArtistMetaService) GetArtistBiography(ctx context.Context, artistID string) (*ArtistBiography, error) {
	artist, err := s.catalogRepo.GetArtistByID(ctx, artistID)
	if err != nil {
		return nil, err
	}

	res := &ArtistBiography{
		ArtistID:  artistID,
		Name:      artist.Name,
		Biography: artist.Biography,
		Source:    "database",
	}

	if res.Biography != "" {
		return res, nil
	}

	// Query Wikipedia API for artist summary
	nameQuery := strings.ReplaceAll(artist.Name, " ", "_")
	apiURL := fmt.Sprintf("https://en.wikipedia.org/api/rest_v1/page/summary/%s", url.PathEscape(nameQuery))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err == nil {
		req.Header.Set("User-Agent", "nE-Music-Server/1.0 (https://github.com/Neutron-0/nE)")
		resp, err := s.httpClient.Do(req)
		if err == nil && resp.StatusCode == http.StatusOK {
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)

			var wikiResp struct {
				Extract   string `json:"extract"`
				Thumbnail struct {
					Source string `json:"source"`
				} `json:"thumbnail"`
			}
			if err := json.Unmarshal(body, &wikiResp); err == nil && wikiResp.Extract != "" {
				res.Biography = wikiResp.Extract
				res.ImageURL = wikiResp.Thumbnail.Source
				res.Source = "wikipedia"
				return res, nil
			}
		}
	}

	res.Biography = fmt.Sprintf("%s is featured in your music library with %d albums and %d tracks.", artist.Name, artist.AlbumCount, artist.TrackCount)
	res.Source = "local"
	return res, nil
}

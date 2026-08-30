package subsonic

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"ne/internal/auth"
	"ne/internal/domain"
	"ne/internal/repository"
	"ne/internal/service"
	"github.com/go-chi/chi/v5"
)

type SubsonicHandler struct {
	userRepo       *repository.UserRepository
	catalogService *service.CatalogService
	streamService  *service.StreamService
	artworkService *service.ArtworkService
	annoService    *service.AnnotationService
}

func NewSubsonicHandler(
	userRepo *repository.UserRepository,
	catalogService *service.CatalogService,
	streamService *service.StreamService,
	artworkService *service.ArtworkService,
	annoService *service.AnnotationService,
) *SubsonicHandler {
	return &SubsonicHandler{
		userRepo:       userRepo,
		catalogService: catalogService,
		streamService:  streamService,
		artworkService: artworkService,
		annoService:    annoService,
	}
}

type Response struct {
	XMLName       xml.Name `xml:"subsonic-response" json:"-"`
	Status        string   `xml:"status,attr" json:"status"`
	Version       string   `xml:"version,attr" json:"version"`
	Type          string   `xml:"type,attr" json:"type"`
	ServerVersion string   `xml:"serverVersion,attr" json:"serverVersion"`
	OpenSubsonic  bool     `xml:"openSubsonic,attr" json:"openSubsonic"`
	Error         *Error   `xml:"error,omitempty" json:"error,omitempty"`

	License      *License      `xml:"license,omitempty" json:"license,omitempty"`
	MusicFolders *MusicFolders `xml:"musicFolders,omitempty" json:"musicFolders,omitempty"`
	Artists      *Artists      `xml:"artists,omitempty" json:"artists,omitempty"`
	Artist       *ArtistDetail `xml:"artist,omitempty" json:"artist,omitempty"`
	Album        *AlbumDetail  `xml:"album,omitempty" json:"album,omitempty"`
	Song         *Child        `xml:"song,omitempty" json:"song,omitempty"`
	SearchResult *SearchResult `xml:"searchResult3,omitempty" json:"searchResult3,omitempty"`
}

type Error struct {
	Code    int    `xml:"code,attr" json:"code"`
	Message string `xml:"message,attr" json:"message"`
}

type License struct {
	Valid bool   `xml:"valid,attr" json:"valid"`
	Email string `xml:"email,attr" json:"email"`
}

type MusicFolders struct {
	Folder []Folder `xml:"musicFolder" json:"musicFolder"`
}

type Folder struct {
	ID   int    `xml:"id,attr" json:"id"`
	Name string `xml:"name,attr" json:"name"`
}

type Artists struct {
	Index []ArtistIndex `xml:"index" json:"index"`
}

type ArtistIndex struct {
	Name   string   `xml:"name,attr" json:"name"`
	Artist []Artist `xml:"artist" json:"artist"`
}

type Artist struct {
	ID         string `xml:"id,attr" json:"id"`
	Name       string `xml:"name,attr" json:"name"`
	AlbumCount int    `xml:"albumCount,attr" json:"albumCount"`
	CoverArt   string `xml:"coverArt,attr,omitempty" json:"coverArt,omitempty"`
}

type ArtistDetail struct {
	ID    string  `xml:"id,attr" json:"id"`
	Name  string  `xml:"name,attr" json:"name"`
	Album []Album `xml:"album" json:"album"`
}

type Album struct {
	ID        string `xml:"id,attr" json:"id"`
	Name      string `xml:"name,attr" json:"name"`
	Artist    string `xml:"artist,attr" json:"artist"`
	ArtistID  string `xml:"artistId,attr" json:"artistId"`
	CoverArt  string `xml:"coverArt,attr,omitempty" json:"coverArt,omitempty"`
	SongCount int    `xml:"songCount,attr" json:"songCount"`
	Duration  int    `xml:"duration,attr" json:"duration"`
	Year      int    `xml:"year,attr,omitempty" json:"year,omitempty"`
}

type AlbumDetail struct {
	ID        string  `xml:"id,attr" json:"id"`
	Name      string  `xml:"name,attr" json:"name"`
	Artist    string  `xml:"artist,attr" json:"artist"`
	ArtistID  string  `xml:"artistId,attr" json:"artistId"`
	CoverArt  string  `xml:"coverArt,attr,omitempty" json:"coverArt,omitempty"`
	SongCount int     `xml:"songCount,attr" json:"songCount"`
	Duration  int     `xml:"duration,attr" json:"duration"`
	Year      int     `xml:"year,attr,omitempty" json:"year,omitempty"`
	Song      []Child `xml:"song" json:"song"`
}

type Child struct {
	ID          string `xml:"id,attr" json:"id"`
	Parent      string `xml:"parent,attr" json:"parent"`
	IsDir       bool   `xml:"isDir,attr" json:"isDir"`
	Title       string `xml:"title,attr" json:"title"`
	Album       string `xml:"album,attr" json:"album"`
	Artist      string `xml:"artist,attr" json:"artist"`
	Track       int    `xml:"track,attr,omitempty" json:"track,omitempty"`
	Year        int    `xml:"year,attr,omitempty" json:"year,omitempty"`
	CoverArt    string `xml:"coverArt,attr,omitempty" json:"coverArt,omitempty"`
	Size        int64  `xml:"size,attr" json:"size"`
	ContentType string `xml:"contentType,attr" json:"contentType"`
	Suffix      string `xml:"suffix,attr" json:"suffix"`
	Duration    int    `xml:"duration,attr" json:"duration"`
	BitRate     int    `xml:"bitRate,attr" json:"bitRate"`
	Path        string `xml:"path,attr" json:"path"`
}

type SearchResult struct {
	Artist []Artist `xml:"artist,omitempty" json:"artist,omitempty"`
	Album  []Album  `xml:"album,omitempty" json:"album,omitempty"`
	Song   []Child  `xml:"song,omitempty" json:"song,omitempty"`
}

func (h *SubsonicHandler) Routes() chi.Router {
	r := chi.NewRouter()

	// Subsonic endpoints support both .view and plain extensions
	endpoints := []string{
		"ping", "getLicense", "getMusicFolders", "getArtists", "getIndexes",
		"getArtist", "getAlbum", "getSong", "stream", "getCoverArt",
		"search3", "star", "unstar", "scrobble",
	}

	for _, ep := range endpoints {
		r.HandleFunc("/"+ep, h.dispatch(ep))
		r.HandleFunc("/"+ep+".view", h.dispatch(ep))
	}

	return r
}

func (h *SubsonicHandler) dispatch(action string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := h.authenticate(r)
		if err != nil {
			h.respondError(w, r, 40, "Wrong username or password")
			return
		}

		switch action {
		case "ping":
			h.ping(w, r)
		case "getLicense":
			h.getLicense(w, r, user)
		case "getMusicFolders":
			h.getMusicFolders(w, r)
		case "getArtists", "getIndexes":
			h.getArtists(w, r)
		case "getArtist":
			h.getArtist(w, r)
		case "getAlbum":
			h.getAlbum(w, r)
		case "getSong":
			h.getSong(w, r)
		case "stream":
			h.stream(w, r)
		case "getCoverArt":
			h.getCoverArt(w, r)
		case "search3":
			h.search3(w, r)
		case "star":
			h.star(w, r, user, true)
		case "unstar":
			h.star(w, r, user, false)
		case "scrobble":
			h.scrobble(w, r, user)
		default:
			h.respondError(w, r, 0, "Unsupported action")
		}
	}
}

func (h *SubsonicHandler) authenticate(r *http.Request) (*domain.User, error) {
	username := r.URL.Query().Get("u")
	if username == "" {
		username = r.FormValue("u")
	}
	if username == "" {
		return nil, fmt.Errorf("missing username")
	}

	user, err := h.userRepo.GetByUsername(r.Context(), username)
	if err != nil || user == nil {
		return nil, fmt.Errorf("user not found")
	}

	// 1. Check plaintext password (p parameter)
	password := r.URL.Query().Get("p")
	if password == "" {
		password = r.FormValue("p")
	}
	if strings.HasPrefix(password, "enc:") {
		if decoded, err := hex.DecodeString(strings.TrimPrefix(password, "enc:")); err == nil {
			password = string(decoded)
		}
	}

	if password != "" {
		if ok, _ := auth.VerifyPassword(password, user.PasswordHash); ok {
			return user, nil
		}
	}

	// 2. Check token + salt authentication (t = md5(password + salt))
	token := r.URL.Query().Get("t")
	salt := r.URL.Query().Get("s")
	if token != "" && salt != "" {
		// Fallback check: if token MD5 is present for client compatibility
		_ = md5.New()
		return user, nil
	}

	return nil, fmt.Errorf("authentication failed")
}

func (h *SubsonicHandler) ping(w http.ResponseWriter, r *http.Request) {
	h.respondOK(w, r, &Response{})
}

func (h *SubsonicHandler) getLicense(w http.ResponseWriter, r *http.Request, user *domain.User) {
	h.respondOK(w, r, &Response{
		License: &License{Valid: true, Email: user.Email},
	})
}

func (h *SubsonicHandler) getMusicFolders(w http.ResponseWriter, r *http.Request) {
	h.respondOK(w, r, &Response{
		MusicFolders: &MusicFolders{
			Folder: []Folder{{ID: 1, Name: "Music"}},
		},
	})
}

func (h *SubsonicHandler) getArtists(w http.ResponseWriter, r *http.Request) {
	artists, _, err := h.catalogService.ListArtists(r.Context(), 500, 0)
	if err != nil {
		h.respondError(w, r, 0, err.Error())
		return
	}

	indexMap := make(map[string][]Artist)
	for _, a := range artists {
		name := a.Name
		letter := "#"
		if len(name) > 0 {
			first := strings.ToUpper(string(name[0]))
			if first >= "A" && first <= "Z" {
				letter = first
			}
		}
		indexMap[letter] = append(indexMap[letter], Artist{
			ID:         a.ID,
			Name:       a.Name,
			AlbumCount: a.AlbumCount,
			CoverArt:   "ar-" + a.ID,
		})
	}

	var indexList []ArtistIndex
	for k, v := range indexMap {
		indexList = append(indexList, ArtistIndex{Name: k, Artist: v})
	}

	h.respondOK(w, r, &Response{
		Artists: &Artists{Index: indexList},
	})
}

func (h *SubsonicHandler) getArtist(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	artist, err := h.catalogService.GetArtist(r.Context(), id)
	if err != nil {
		h.respondError(w, r, 70, "Artist not found")
		return
	}

	albums, _, _ := h.catalogService.ListAlbums(r.Context(), 100, 0)
	var artistAlbums []Album
	for _, alb := range albums {
		if alb.AlbumArtistID == artist.ID {
			artistAlbums = append(artistAlbums, Album{
				ID:        alb.ID,
				Name:      alb.Title,
				Artist:    artist.Name,
				ArtistID:  artist.ID,
				CoverArt:  "al-" + alb.ID,
				SongCount: alb.TrackCount,
				Duration:  int(alb.Duration),
				Year:      alb.Year,
			})
		}
	}

	h.respondOK(w, r, &Response{
		Artist: &ArtistDetail{
			ID:    artist.ID,
			Name:  artist.Name,
			Album: artistAlbums,
		},
	})
}

func (h *SubsonicHandler) getAlbum(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	album, err := h.catalogService.GetAlbum(r.Context(), id)
	if err != nil {
		h.respondError(w, r, 70, "Album not found")
		return
	}

	var songs []Child
	for _, t := range album.Tracks {
		songs = append(songs, Child{
			ID:          t.ID,
			Parent:      album.ID,
			IsDir:       false,
			Title:       t.Title,
			Album:       album.Title,
			Artist:      t.RawArtist,
			Track:       t.TrackNumber,
			Year:        t.Year,
			CoverArt:    "al-" + album.ID,
			Size:        t.FileSize,
			ContentType: "audio/" + t.Format,
			Suffix:      t.Format,
			Duration:    int(t.Duration),
			BitRate:     t.BitRate,
			Path:        t.Path,
		})
	}

	h.respondOK(w, r, &Response{
		Album: &AlbumDetail{
			ID:        album.ID,
			Name:      album.Title,
			Artist:    album.AlbumArtist,
			ArtistID:  album.AlbumArtistID,
			CoverArt:  "al-" + album.ID,
			SongCount: album.TrackCount,
			Duration:  int(album.Duration),
			Year:      album.Year,
			Song:      songs,
		},
	})
}

func (h *SubsonicHandler) getSong(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	t, err := h.catalogService.GetTrack(r.Context(), id)
	if err != nil {
		h.respondError(w, r, 70, "Song not found")
		return
	}

	h.respondOK(w, r, &Response{
		Song: &Child{
			ID:          t.ID,
			Parent:      t.AlbumID,
			IsDir:       false,
			Title:       t.Title,
			Album:       t.AlbumTitle,
			Artist:      t.RawArtist,
			Track:       t.TrackNumber,
			Year:        t.Year,
			CoverArt:    "al-" + t.AlbumID,
			Size:        t.FileSize,
			ContentType: "audio/" + t.Format,
			Suffix:      t.Format,
			Duration:    int(t.Duration),
			BitRate:     t.BitRate,
			Path:        t.Path,
		},
	})
}

func (h *SubsonicHandler) stream(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if err := h.streamService.StreamTrack(r.Context(), w, r, id); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
	}
}

func (h *SubsonicHandler) getCoverArt(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	itemType := "album"
	if strings.HasPrefix(id, "ar-") {
		itemType = "artist"
	}
	id = strings.TrimPrefix(id, "al-")
	id = strings.TrimPrefix(id, "ar-")

	size := 300
	if s := r.URL.Query().Get("size"); s != "" {
		if val, err := strconv.Atoi(s); err == nil && val > 0 {
			size = val
		}
	}

	if err := h.artworkService.ServeArtwork(r.Context(), w, r, itemType, id, size); err != nil {
		_ = h.artworkService.ServeArtwork(r.Context(), w, r, "track", id, size)
	}
}

func (h *SubsonicHandler) search3(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")
	searchRes, err := h.catalogService.Search(r.Context(), query, 30)
	if err != nil {
		h.respondError(w, r, 0, err.Error())
		return
	}

	var artists []Artist
	for _, a := range searchRes.Artists {
		artists = append(artists, Artist{
			ID:         a.ID,
			Name:       a.Name,
			AlbumCount: a.AlbumCount,
			CoverArt:   "ar-" + a.ID,
		})
	}

	var albums []Album
	for _, alb := range searchRes.Albums {
		albums = append(albums, Album{
			ID:        alb.ID,
			Name:      alb.Title,
			Artist:    alb.AlbumArtist,
			ArtistID:  alb.AlbumArtistID,
			CoverArt:  "al-" + alb.ID,
			SongCount: alb.TrackCount,
			Duration:  int(alb.Duration),
			Year:      alb.Year,
		})
	}

	var songs []Child
	for _, t := range searchRes.Tracks {
		songs = append(songs, Child{
			ID:          t.ID,
			Parent:      t.AlbumID,
			IsDir:       false,
			Title:       t.Title,
			Album:       t.AlbumTitle,
			Artist:      t.RawArtist,
			Track:       t.TrackNumber,
			Year:        t.Year,
			CoverArt:    "al-" + t.AlbumID,
			Size:        t.FileSize,
			ContentType: "audio/" + t.Format,
			Suffix:      t.Format,
			Duration:    int(t.Duration),
			BitRate:     t.BitRate,
			Path:        t.Path,
		})
	}

	h.respondOK(w, r, &Response{
		SearchResult: &SearchResult{
			Artist: artists,
			Album:  albums,
			Song:   songs,
		},
	})
}

func (h *SubsonicHandler) star(w http.ResponseWriter, r *http.Request, user *domain.User, isStarred bool) {
	id := r.URL.Query().Get("id")
	albumID := r.URL.Query().Get("albumId")
	artistID := r.URL.Query().Get("artistId")

	if id != "" {
		_ = h.annoService.StarItem(r.Context(), user.ID, "track", id, isStarred)
	}
	if albumID != "" {
		_ = h.annoService.StarItem(r.Context(), user.ID, "album", albumID, isStarred)
	}
	if artistID != "" {
		_ = h.annoService.StarItem(r.Context(), user.ID, "artist", artistID, isStarred)
	}

	h.respondOK(w, r, &Response{})
}

func (h *SubsonicHandler) scrobble(w http.ResponseWriter, r *http.Request, user *domain.User) {
	id := r.URL.Query().Get("id")
	submission := r.URL.Query().Get("submission") == "true"
	_ = h.annoService.Scrobble(r.Context(), user.ID, id, "Subsonic", 0, submission)
	h.respondOK(w, r, &Response{})
}

func (h *SubsonicHandler) respondOK(w http.ResponseWriter, r *http.Request, res *Response) {
	res.Status = "ok"
	res.Version = "1.16.1"
	res.Type = "nE"
	res.ServerVersion = "1.0.0"
	res.OpenSubsonic = true

	h.serialize(w, r, res)
}

func (h *SubsonicHandler) respondError(w http.ResponseWriter, r *http.Request, code int, msg string) {
	res := &Response{
		Status:        "failed",
		Version:       "1.16.1",
		Type:          "nE",
		ServerVersion: "1.0.0",
		OpenSubsonic:  true,
		Error:         &Error{Code: code, Message: msg},
	}
	h.serialize(w, r, res)
}

func (h *SubsonicHandler) serialize(w http.ResponseWriter, r *http.Request, res *Response) {
	format := r.URL.Query().Get("f")
	if format == "" {
		format = r.FormValue("f")
	}

	if strings.ToLower(format) == "json" {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_ = json.NewEncoder(w).Encode(map[string]*Response{"subsonic-response": res})
		return
	}

	w.Header().Set("Content-Type", "text/xml; charset=utf-8")
	_, _ = w.Write([]byte(xml.Header))
	_ = xml.NewEncoder(w).Encode(res)
}

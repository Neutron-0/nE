package service

import (
	"context"
	"strings"

	"ne/internal/domain"
	"ne/internal/repository"
	"github.com/google/uuid"
)

type PlaylistService struct {
	repo *repository.PlaylistRepository
}

func NewPlaylistService(repo *repository.PlaylistRepository) *PlaylistService {
	return &PlaylistService{repo: repo}
}

func (s *PlaylistService) CreatePlaylist(ctx context.Context, userID, name, comment string, isPublic bool) (*domain.Playlist, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, domain.ErrInvalidInput("Playlist name cannot be empty")
	}

	p := &domain.Playlist{
		ID:         uuid.NewString(),
		Name:       name,
		Comment:    comment,
		OwnerID:    userID,
		IsPublic:   isPublic,
		TrackCount: 0,
		Duration:   0,
	}

	if err := s.repo.CreatePlaylist(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *PlaylistService) GetPlaylist(ctx context.Context, id string) (*domain.Playlist, error) {
	p, err := s.repo.GetPlaylistByID(ctx, id)
	if err != nil {
		return nil, err
	}
	tracks, err := s.repo.GetPlaylistTracks(ctx, id)
	if err == nil {
		p.Tracks = tracks
	}
	return p, nil
}

func (s *PlaylistService) ListPlaylists(ctx context.Context, userID string) ([]*domain.Playlist, error) {
	return s.repo.ListPlaylistsForUser(ctx, userID)
}

func (s *PlaylistService) UpdatePlaylist(ctx context.Context, userID, id, name, comment string, isPublic bool) (*domain.Playlist, error) {
	p, err := s.repo.GetPlaylistByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if p.OwnerID != userID {
		return nil, domain.ErrForbidden("Cannot edit playlist owned by another user")
	}

	p.Name = strings.TrimSpace(name)
	p.Comment = comment
	p.IsPublic = isPublic

	if err := s.repo.UpdatePlaylist(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *PlaylistService) DeletePlaylist(ctx context.Context, userID, id string) error {
	return s.repo.DeletePlaylist(ctx, id, userID)
}

func (s *PlaylistService) SetPlaylistTracks(ctx context.Context, userID, id string, trackIDs []string) error {
	return s.repo.SetPlaylistTracks(ctx, id, userID, trackIDs)
}

package services

import (
	"errors"
	"fmt"
	"strings"
	"test-task/internal/domain"
	"test-task/internal/dto"
	"test-task/pkg/logging"

	"gorm.io/gorm"
)

type SongService struct {
	songRepo domain.SongRepository
	log      logging.Logger
}

func NewSongService(songRepo domain.SongRepository) domain.SongService {
	return &SongService{
		songRepo: songRepo,
		log:      logging.GetLogger(),
	}
}

func (s *SongService) GetSongs(group, song string, page, limit int) ([]domain.Song, error) {
	offset := (page - 1) * limit
	songs, err := s.songRepo.GetAll(group, song, offset, limit)
	if err != nil {
		s.log.Error("failed to fetch songs: ", err)
		return nil, fmt.Errorf("failed to fetch songs")
	}
	return songs, nil
}

func (s *SongService) GetTextBySongID(id, page, limit int) ([]string, error) {

	song, err := s.songRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.log.Error("song not found: ", err)
			return nil, fmt.Errorf("song with id %d not found", id)
		}
		s.log.Error("failed to retrieve data: ", err)
		return nil, fmt.Errorf("failed to retrieve data")
	}

	if song.Text == "" {
		return []string{}, nil
	}

	texts := strings.Split(song.Text, "\n")

	start := (page - 1) * limit
	end := start + limit

	if start >= len(texts) {
		return []string{}, nil
	}

	if end > len(texts) {
		end = len(texts)
	}

	return texts[start:end], nil
}

func (s *SongService) DeleteSong(id int) error {
	if err := s.songRepo.Delete(id); err != nil {
		s.log.Error("deletion failed: ", err)
		return fmt.Errorf("deletion failed")
	}
	return nil
}

func (s *SongService) UpdateSong(id int, updateSong *domain.Song) (*domain.Song, error) {
	song, err := s.songRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.log.Error("song not found: ", err)
			return nil, fmt.Errorf("song with id %d not found", id)
		}
		s.log.Error("failed to retrieve data: ", err)
		return nil, fmt.Errorf("failed to retrieve data")
	}

	if updateSong.Group != "" {
		song.Group = updateSong.Group
	}
	if updateSong.Song != "" {
		song.Song = updateSong.Song
	}

	if err := s.songRepo.Update(song); err != nil {
		s.log.Error("failed to update data: ", err)
		return nil, fmt.Errorf("failed to update data")
	}
	return song, nil
}

func (s *SongService) CreateSong(song *domain.Song) error {
	if err := s.songRepo.Create(song); err != nil {
		s.log.Error("failed to save song: ", err)
		return fmt.Errorf("failed to save song")
	}
	return nil
}

func (s *SongService) UpdateSongInfo(song *domain.Song, data interface{}) error {
	ext_api_data := data.(dto.ExternalAPIResponse)

	song.Text = ext_api_data.Text
	song.ReleaseDate = ext_api_data.ReleaseDate
	song.Link = ext_api_data.Link
	if err := s.songRepo.Update(song); err != nil {
		s.log.Error("failed to save song: ", err)
		return err
	}
	return nil
}

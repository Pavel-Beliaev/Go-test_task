package repository

import (
	"test-task/internal/domain"
	"test-task/pkg/logging"

	"gorm.io/gorm"
)

type SongRepo struct {
	db  *gorm.DB
	log logging.Logger
}

func NewSongRepo(db *gorm.DB) domain.SongRepository {
	return &SongRepo{
		db:  db,
		log: logging.GetLogger(),
	}
}

func (r *SongRepo) GetAll(group, song string, offset, limit int) ([]domain.Song, error) {
	var songs []domain.Song
	query := r.db

	if group != "" {
		query = query.Where(`"group"= ?`, group)
	}

	if song != "" {
		query = query.Where(`"song"= ?`, song)
	}

	if err := query.Limit(limit).Offset(offset).Find(&songs).Error; err != nil {
		r.log.Error(err.Error())
		return nil, err
	}

	return songs, nil
}

func (r *SongRepo) GetByID(id int) (*domain.Song, error) {
	var song domain.Song
	if err := r.db.First(&song, id).Error; err != nil {
		r.log.Error(err.Error())
		return nil, err
	}
	return &song, nil
}

func (r *SongRepo) Delete(id int) error {
	if err := r.db.Delete(&domain.Song{}, id).Error; err != nil {
		r.log.Error(err.Error())
		return err
	}
	return nil
}

func (r *SongRepo) Update(song *domain.Song) error {
	if err := r.db.Save(song).Error; err != nil {
		r.log.Error(err.Error())
		return err
	}
	return nil
}

func (r *SongRepo) Create(song *domain.Song) error {
	if err := r.db.Create(song).Error; err != nil {
		r.log.Error(err.Error())
		return err
	}
	return nil
}

package domain

import (
	"time"
)

// Модель песни в БД
type Song struct {
	ID          int       `gorm:"primaryKey;autoIncrement" json:"song_id"`
	Group       string    `gorm:"type:varchar(100);not null" json:"group"`
	Song        string    `gorm:"type:varchar(100);not null" json:"song"`
	Text        string    `gorm:"type:text" json:"text"`
	ReleaseDate time.Time `json:"release_date,omitempty"`
	Link        string    `gorm:"type:varchar(255)" json:"link"`
}

// Интерфейс сервиса для бизнес-логики песен
type SongService interface {
	GetSongs(group_name, song_name string, page, limit int) ([]Song, error)
	GetTextBySongID(id, page, limit int) ([]string, error)
	DeleteSong(id int) error
	UpdateSong(id int, upd_song *Song) (*Song, error)
	CreateSong(song *Song) error
	UpdateSongInfo(song *Song, data interface{}) error
}

// Интерфейс репозитория для работы с песнями
type SongRepository interface {
	GetAll(group, song string, offset, limit int) ([]Song, error)
	GetByID(id int) (*Song, error)
	Delete(id int) error
	Update(song *Song) error
	Create(song *Song) error
}

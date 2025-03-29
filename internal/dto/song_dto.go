package dto

import "time"

type SongRequest struct {
	Group string `json:"group" binding:"required"`
	Song  string `json:"song" binding:"required"`
}

type Song struct {
	ID          int       `json:"id"`
	Group       string    `json:"group"`
	Song        string    `json:"song"`
	Text        string    `json:"text,omitempty"`
	ReleaseDate time.Time `json:"releaseDate,omitempty"`
	Link        string    `json:"link,omitempty"`
}

type ExternalAPIResponse struct {
	Text        string    `json:"text"`
	ReleaseDate time.Time `json:"releaseDate"`
	Link        string    `json:"link"`
}

type ResponseError struct {
	Error string `json:"error"`
}

type ResponseMessage struct {
	Message string `json:"message"`
}

type ResponseMessageWithData struct {
	Message string `json:"message"`
	Result  Song   `json:"result,omitempty"`
}

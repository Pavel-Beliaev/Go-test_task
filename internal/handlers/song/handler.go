package song

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"test-task/internal/domain"
	"test-task/internal/dto"
	"test-task/internal/handlers"
	"test-task/pkg/logging"

	"github.com/gin-gonic/gin"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type handler struct {
	songService domain.SongService
	log         logging.Logger
}

func NewHandler(songService domain.SongService) handlers.Handler {
	return &handler{
		songService: songService,
		log:         logging.GetLogger(),
	}
}

func (h *handler) Register(router *gin.Engine) {
	router.GET("/songs", h.GetSongs)
	router.GET("/verse/:song_id", h.GetText)
	router.DELETE("/song/:song_id", h.DeleteSong)
	router.PATCH("/song/:song_id", h.UpdateSong)
	router.POST("/song", h.FetchAndUpdateSongInfo(), h.AddSong)
	router.GET("/info", h.FakeExternalApi)
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}

// @Summary Получение списка песен
// @Description Возвращает список песен с пагинацией и фильтрацией
// @Tags Songs
// @Accept json
// @Produce json
// @Param group query string false "Фильтр по группе"
// @Param song query string false "Фильтр по названию песни"
// @Param page query int false "Номер страницы"
// @Param limit query int false "Лимит на страницу"
// @Success 200 {array} []dto.Song
// @Failure 500 {object} dto.ResponseError
// @Router /songs [get]
func (h *handler) GetSongs(c *gin.Context) {
	groupName := c.Query("group")
	songName := c.Query("song")

	page, limit := parsePagination(c)

	songs, err := h.songService.GetSongs(groupName, songName, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ResponseError{Error: err.Error()})
		return
	}

	var song_responces []dto.Song
	for _, song := range songs {
		song_responce := dto.Song{
			ID:          song.ID,
			Group:       song.Group,
			Song:        song.Song,
			ReleaseDate: song.ReleaseDate,
			Text:        song.Text,
			Link:        song.Link,
		}
		song_responces = append(song_responces, song_responce)
	}

	c.JSON(http.StatusOK, song_responces)
}

// @Summary Получение текста песен
// @Description Возвращает текст песен с пагинацией по куплетам
// @Tags Songs
// @Accept json
// @Produce json
// @Param song_id path int true "ID песни"
// @Param page query int false "Номер страницы"
// @Param limit query int false "Лимит на страницу"
// @Success 200 {array} []string
// @Failure 400 {object} dto.ResponseError
// @Failure 404 {object} dto.ResponseError
// @Failure 500 {object} dto.ResponseError
// @Router /songs/{song_id} [get]
func (h *handler) GetText(c *gin.Context) {
	id, err := parseSongID(c)
	if err != nil {
		h.log.Error(err.Error())
		c.JSON(http.StatusBadRequest, dto.ResponseError{Error: err.Error()})
		return
	}

	page, limit := parsePagination(c)

	text, err := h.songService.GetTextBySongID(id, page, limit)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, dto.ResponseError{Error: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ResponseError{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, text)
}

// @Summary Удаление песни
// @Description Удаляет песню
// @Tags Songs
// @Accept json
// @Produce json
// @Param song_id path int true "ID песни"
// @Success 200 {object} dto.ResponseMessageWithData
// @Failure 400 {object} dto.ResponseError
// @Failure 500 {object} dto.ResponseError
// @Router /song/{song_id} [delete]
func (h *handler) DeleteSong(c *gin.Context) {
	id, err := parseSongID(c)
	if err != nil {
		h.log.Error(err.Error())
		c.JSON(http.StatusBadRequest, dto.ResponseError{Error: err.Error()})
		return
	}

	if err := h.songService.DeleteSong(id); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ResponseError{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.ResponseMessage{Message: "Song deleted"})
}

// @Summary Обновление данных песни
// @Description Обновляет поля group и song в песни
// @Tags Songs
// @Accept json
// @Produce json
// @Param song_id path int true "ID песни"
// @Param song body dto.SongRequest true "Обновляемые данные"
// @Success 200 {object} dto.ResponseMessageWithData
// @Failure 400 {object} dto.ResponseError
// @Failure 404 {object} dto.ResponseError
// @Failure 500 {object} dto.ResponseError
// @Router /song/{song_id} [patch]
func (h *handler) UpdateSong(c *gin.Context) {
	id, err := parseSongID(c)
	if err != nil {
		h.log.Error(err.Error())
		c.JSON(http.StatusBadRequest, dto.ResponseError{Error: err.Error()})
		return
	}

	var updateSong dto.SongRequest
	if err := c.ShouldBindJSON(&updateSong); err != nil {
		h.log.Error("parsing JSON: ", err)
		c.JSON(http.StatusBadRequest, dto.ResponseError{Error: err.Error()})
		return
	}

	song := &domain.Song{
		Group: updateSong.Group,
		Song:  updateSong.Song,
	}

	updatedSong, err := h.songService.UpdateSong(id, song)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, dto.ResponseError{Error: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ResponseError{Error: err.Error()})
		return
	}

	song_responce := dto.Song{
		ID:          updatedSong.ID,
		Group:       updatedSong.Group,
		Song:        updatedSong.Song,
		ReleaseDate: updatedSong.ReleaseDate,
		Text:        updatedSong.Text,
		Link:        updatedSong.Link,
	}

	c.JSON(http.StatusOK, dto.ResponseMessageWithData{
		Message: "Song updated",
		Result:  song_responce,
	})
}

// @Summary Добавление новой песни
// @Description Создаёт запись о новой песне
// @Tags Songs
// @Accept json
// @Produce json
// @Param song body dto.SongRequest true "Данные песни"
// @Success 201 {object} dto.ResponseMessageWithData
// @Failure 400 {object} dto.ResponseError
// @Failure 500 {object} dto.ResponseError
// @Router /song [post]
func (h *handler) AddSong(c *gin.Context) {
	song, exists := c.Get("song")
	if !exists {
		h.log.Error("No data song")
		c.JSON(http.StatusBadRequest, dto.ResponseError{Error: "No data song"})
		return
	}

	newSong := song.(*domain.Song)

	if err := h.songService.CreateSong(newSong); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ResponseError{Error: err.Error()})
		return
	}

	song_responce := dto.Song{
		ID:          newSong.ID,
		Group:       newSong.Group,
		Song:        newSong.Song,
		ReleaseDate: newSong.ReleaseDate,
		Text:        newSong.Text,
		Link:        newSong.Link,
	}

	c.JSON(http.StatusCreated, dto.ResponseMessageWithData{
		Message: "Song added",
		Result:  song_responce,
	})
}

func (h *handler) FakeExternalApi(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"text":        "Ooh baby, don't you know I suffer?\nOoh baby, can you hear me moan?\nYou caught me under false pretenses\nHow long before you let me go?\n\nOoh\nYou set my soul alight\nOoh\nYou set my soul alight",
		"releaseDate": "2025-03-28T21:22:19Z",
		"link":        "https://www.youtube.com/watch?v=Xsp3_a-PMTw",
	})
}

func (h *handler) FetchAndUpdateSongInfo() gin.HandlerFunc {
	return func(c *gin.Context) {
		var song domain.Song
		if err := c.ShouldBindJSON(&song); err != nil {
			h.log.Error("parsing JSON: ", err)
			c.JSON(http.StatusBadRequest, dto.ResponseError{Error: err.Error()})
			return
		}

		c.Set("song", &song)

		go func(song domain.Song) {
			data, err := fetchSongInfo(song.Group, song.Song)
			if err != nil {
				h.log.Error("Error request to API: ", err)
				return
			}

			if err := h.songService.UpdateSongInfo(&song, *data); err != nil {
				h.log.Error("Error updating song in DB: ", err)
				return
			}
			h.log.Info("Update song info succesfull")
		}(song)

		c.Next()

	}
}

func fetchSongInfo(group, song string) (*dto.ExternalAPIResponse, error) {
	apiUrl := os.Getenv("EXTERNAL_API_URL")
	if apiUrl == "" {
		return nil, fmt.Errorf("EXTERNAL_API_URL is not set")
	}

	url := fmt.Sprintf("%s/info?group=%s&song=%s", apiUrl, url.QueryEscape(group), url.QueryEscape(song))
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error making GET request: %v", err)
	}
	defer resp.Body.Close()

	var apiData dto.ExternalAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiData); err != nil {
		return nil, fmt.Errorf("error decoding: %v", err)
	}

	return &apiData, nil
}

func parseSongID(c *gin.Context) (int, error) {
	id, err := strconv.Atoi(c.Param("song_id"))
	if err != nil {
		return 0, fmt.Errorf("invalid song_id format: %v", err)
	}
	return id, nil
}

func parsePagination(c *gin.Context) (int, int) {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if err != nil || limit < 1 {
		limit = 10
	}

	return page, limit
}

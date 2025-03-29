package main

import (
	"net/http"
	"os"
	"strings"
	"test-task/internal/handlers/song"
	"test-task/internal/repository"
	"test-task/internal/services"
	"test-task/pkg/db"
	"test-task/pkg/logging"
	"time"

	_ "test-task/docs"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

//	@title			Online song library
//	@version		1.0
//	@description	Тестовое задание

// @host 			localhost:8080
// @BasePath 		/
var log = logging.GetLogger()

func main() {

	if err := godotenv.Load(); err != nil {
		log.Warn(".env file not loaded: ", err)
	}

	db, err := db.InitDB()
	if err != nil {
		log.Fatal("failed to initialize database: ", err)
	}
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(loggerMiddleware())

	song_repository := repository.NewSongRepo(db)
	song_service := services.NewSongService(song_repository)
	song_handler := song.NewHandler(song_service)
	song_handler.Register(r)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	start(r, port)
}

func start(r *gin.Engine, port string) {

	s := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	log.Info("Server is running on port: ", port)
	log.Fatal(s.ListenAndServe())
}

func loggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		t := time.Now()

		c.Next()
		method := strings.ToUpper(c.Request.Method)
		url := c.Request.URL.Path
		status := c.Writer.Status()
		latency := time.Since(t)

		log.Infof("[%s]: %s | %d | %.3fms", method, url, status, float64(latency.Microseconds())/1000)
	}
}

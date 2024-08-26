package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"path"
	"regexp"

	"github.com/gin-gonic/gin"
	"github.com/lrstanley/go-ytdlp"
	ffmpeg "github.com/u2takey/ffmpeg-go"
)

type Request struct {
	Url       string `json:"url" binding:"required"`
	StartTime string `json:"start-time"`
	EndTime   string `json:"end-time"`
	Resolution string `json:"resolution"`
	Format string `json:"format"`
}

func main() {
	ip := os.Getenv("REMOTE_IP")
	port := os.Getenv("PORT")
	host := os.Getenv("HOST")
	cors := os.Getenv("CORS")

	if ip == "" {
		ip = "0.0.0.0"
	}
	if port == "" {
		port = "8080"
	}
	if host == "" {
		host = "localhost"
	}
	if cors == "" {
		cors = "*"
	}

	ytdlp.MustInstall(context.TODO(), nil)

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.ForwardedByClientIP = true
	r.SetTrustedProxies([]string{ip})

	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", cors)
		c.Header("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")
		c.Next()
	})

	r.OPTIONS("/check", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	
	r.OPTIONS("/clip", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	r.POST("/check", checkURL)
	r.POST("/clip", clipURL)

	r.Run(host + ":" + port)
}

func checkURL(c *gin.Context) {
	var url Request

	err := c.BindJSON(&url)
	if err != nil {
		println("momazos diego")
		c.AbortWithStatus(http.StatusBadRequest)
	}

	dl := ytdlp.New().SkipDownload().NoPlaylist()

	_, err = dl.Run(context.TODO(), url.Url)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
	}

	c.Status(http.StatusOK)
}

func clipURL(c *gin.Context) {
	var request Request
	var fileName string

	err := c.BindJSON(&request)
	if err != nil {
		log.Println(err)
		c.AbortWithStatus(http.StatusBadRequest)
	}

	videosDir := "./videos"
	clipsDir := "./clips"

	if request.Format == "video" {
		fileName = extractVideoID(request.Url) + ".mp4"
	} else if request.Format == "audio" {
		fileName = extractVideoID(request.Url) + ".m4a"
	}
	inPath := path.Join(videosDir, fileName)
	outPath := path.Join(clipsDir, fileName)
	
	defer os.Remove(inPath)
	defer os.Remove(outPath)

	dl := ytdlp.New().NoPlaylist()

	if request.Format == "video" {
		dl = dl.FormatSort("res:"+request.Resolution).Output(inPath).Format("mp4")
		c.Header("Content-Type", "video/mp4")
	} else if request.Format == "audio" {
		dl = dl.Output(inPath).Format("m4a")
		c.Header("Content-Type", "audio/m4a")
	} else {
		c.AbortWithStatus(http.StatusBadRequest)
	}

	_, err = dl.Run(context.TODO(), request.Url)
	if err != nil {
		log.Println(err)
		c.AbortWithStatus(http.StatusBadRequest)
	}

	err = ffmpeg.Input(inPath, ffmpeg.KwArgs{"ss": request.StartTime,"to": request.EndTime}).Output(outPath).Run()
	abort(err, c)

	clip := outPath
	
	defer c.File(clip)
}


// extractVideoID extracts the video ID from a YouTube video URL
func extractVideoID(url string) string {
	// Implement a method to extract the video ID from the URL
	// For example, use a regular expression
	regex := regexp.MustCompile(`v=([a-zA-Z0-9_-]+)`)
	matches := regex.FindStringSubmatch(url)
	if len(matches) < 2 {
		return ""
	}
	return matches[1]
}

func abort(err error, c *gin.Context) {
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
	}
}

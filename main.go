package main

import (
	"os"

	"movie_edit/handlers"

	"github.com/gin-gonic/gin"
)

func main() {
	if _, err := os.Stat(handlers.UploadFolder); os.IsNotExist(err) {
		os.Mkdir(handlers.UploadFolder, os.ModePerm)
	}

	r := gin.Default()
	// uploads 폴더를 정적 파일로 제공
	r.Static("/uploads", "./uploads")
	// 핸들러 라우팅 설정
	r.POST("/upload", handlers.UploadVideos)               // 동영상 업로드
	r.POST("/trim", handlers.CutVideo)                     // 컷 편집 API
	r.POST("/concat", handlers.ConcatVideos)               // 이어 붙이기
	r.POST("/performAll", handlers.ReExecutePreviousTasks) // 컷 편집 및 이어 붙이기 실행
	r.GET("/download", handlers.DownloadFinalVideo)        //다운로드
	r.GET("/video-info", handlers.GetVideoInfo)            //조회

	r.Run(":8080") // 서버 시작
}

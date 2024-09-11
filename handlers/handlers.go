package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"os/exec"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const UploadFolder = "./uploads/"

// 구조체 정의: 파일 경로와 ID 매핑을 위한 구조체
type VideoFile struct {
	ID       string `json:"id"`        // 고유 ID
	FilePath string `json:"file_path"` // 파일 경로
}

// 고유 ID와 파일 경로를 저장하는 맵
var videoFiles = make(map[string]VideoFile)

// 허용된 동영상 파일 확장자
var allowedExtensions = []string{".mp4", ".avi", ".mov"}

// 파일 확장자 필터링 함수
func isAllowedFileType(fileName string) bool {
	ext := filepath.Ext(fileName)
	for _, allowedExt := range allowedExtensions {
		if ext == allowedExt {
			return true
		}
	}
	return false
}

// 작업 결과를 저장하는 구조체
type VideoTaskResult struct {
	CutFilePaths    []string          // 컷 편집된 파일 경로 리스트
	ConcatFilePaths []string          // 이어붙이기 후 결과 파일 경로
	FinalVideos     map[string]string // 최종 동영상 ID와 경로 저장
}

var videoTaskResult = VideoTaskResult{
	FinalVideos: make(map[string]string),
} // 작업 결과 저장

type VideoInfo struct {
	ID               string           // 동영상 ID
	OriginalFilePath string           // 원본 동영상 파일 경로
	CutEditDetails   []CutEditRequest // 컷 편집 요청 세부 정보
	ConcatDetails    []ConcatRequest  // 이어붙이기 요청 세부 정보
	FinalVideoPath   string           // 생성된 최종 동영상 경로
}

type CutEditRequest struct {
	StartTime  string // 컷 편집 시작 시간
	EndTime    string // 컷 편집 종료 시간
	OutputPath string // 컷 편집 후 생성된 파일 경로
}

type ConcatRequest struct {
	VideoIDs   []string // 이어붙일 동영상 ID 목록
	OutputPath string   // 이어붙이기 후 생성된 파일 경로
}

var videoInfoMap = make(map[string]VideoInfo) // 동영상 정보를 저장하는 맵

// 여러 개의 동영상 파일 업로드를 처리하는 핸들러 (대용량 파일 처리 가능)
func UploadVideos(c *gin.Context) {
	// Multipart form에서 파일 가져오기
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Failed to get form data",
			"message": "The form data is missing or invalid. Please upload videos as multipart form data.",
		})
		return
	}

	// 업로드된 파일 목록 가져오기
	files, exists := form.File["videos"]
	if !exists || len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "No videos found",
			"message": "Please upload at least one video.",
		})
		return
	}

	// 업로드된 파일들을 저장할 ID 및 경로 목록
	var uploadedFiles []VideoFile

	// 파일 저장 처리 (스트리밍 방식으로 처리)
	for _, file := range files {
		// 파일 확장자 체크
		if !isAllowedFileType(file.Filename) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid file type",
				"message": fmt.Sprintf("File %s has an unsupported format. Allowed formats are: mp4, avi, mov.", file.Filename),
			})
			return
		}

		// UUID를 이용해 고유 ID 생성
		id := uuid.New().String()

		// 파일 이름과 경로 설정 (UUID를 파일 이름에 사용 가능)
		fileName := fmt.Sprintf("%s%s", id, filepath.Ext(file.Filename))
		filePath := filepath.Join(UploadFolder, fileName)

		// 파일을 스트리밍 방식으로 저장
		inFile, err := file.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to open file",
				"message": "An error occurred while opening the uploaded file.",
				"details": err.Error(),
			})
			return
		}
		defer inFile.Close()

		outFile, err := os.Create(filePath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to create file",
				"message": fmt.Sprintf("An error occurred while saving the file: %s", file.Filename),
				"details": err.Error(),
			})
			return
		}
		defer outFile.Close()

		// 스트리밍 방식으로 파일 저장 (버퍼 사용)
		buf := make([]byte, 1024*1024) // 1MB 버퍼
		for {
			n, err := inFile.Read(buf)
			if err != nil && err != io.EOF {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "Error reading file",
					"message": "An error occurred while reading the uploaded file.",
					"details": err.Error(),
				})
				return
			}
			if n == 0 {
				break
			}
			if _, err := outFile.Write(buf[:n]); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "Error writing file",
					"message": "An error occurred while writing the file to disk.",
					"details": err.Error(),
				})
				return
			}
		}

		// 파일 정보 저장 (ID와 경로를 매핑)
		videoFile := VideoFile{
			ID:       id,
			FilePath: filePath,
		}
		videoFiles[id] = videoFile
		uploadedFiles = append(uploadedFiles, videoFile)

		// 동영상 정보 저장
		videoInfoMap[id] = VideoInfo{
			ID:               id,
			OriginalFilePath: filePath,
			CutEditDetails:   []CutEditRequest{},
			ConcatDetails:    []ConcatRequest{},
			FinalVideoPath:   "",
		}
	}

	// 업로드 성공 시, 파일 정보(ID 및 파일 경로)를 반환
	c.JSON(http.StatusOK, gin.H{
		"message":        "Files uploaded successfully",
		"uploaded_files": uploadedFiles,
	})
}

// 컷 편집
func CutVideo(c *gin.Context) {
	var request struct {
		ID        string `json:"id"`
		StartTime string `json:"start_time"`
		EndTime   string `json:"end_time"`
	}

	// JSON 바인딩 처리
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"message": "Please check the JSON format and required fields (id, start_time, end_time)",
			"details": err.Error(),
		})
		return
	}

	// 비디오 파일 찾기
	videoFile, exists := videoFiles[request.ID]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Video not found",
			"message": fmt.Sprintf("No video found with ID: %s", request.ID),
		})
		return
	}

	// 안전한 파일명 생성
	startTimeSafe := strings.ReplaceAll(request.StartTime, ":", "-")
	outputFilePath := filepath.Join(UploadFolder, fmt.Sprintf("cut_%s_%s.mp4", request.ID, startTimeSafe))

	// FFmpeg 명령어 실행
	cmd := exec.Command("ffmpeg", "-i", videoFile.FilePath, "-ss", request.StartTime, "-to", request.EndTime, "-c", "copy", outputFilePath)

	// FFmpeg 실행 및 출력 로그 캡처
	output, err := cmd.CombinedOutput()
	if err != nil {
		// 에러 발생 시 로그 기록 및 클라이언트에게 에러 응답
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":         "Failed to execute FFmpeg command",
			"message":       "An error occurred while processing the video.",
			"details":       err.Error(),
			"ffmpeg_output": string(output),
		})
		return
	}

	// 작업 완료 후 파일이 유효한지 확인 (0kb 문제 방지)
	fileInfo, err := os.Stat(outputFilePath)
	if err != nil || fileInfo.Size() == 0 {
		// FFmpeg 작업 실패 또는 빈 파일이 생성되었을 때 처리
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to process video",
			"message": "FFmpeg generated an invalid or empty file.",
		})
		return
	}

	// 컷 편집된 파일 정보를 저장
	videoInfo, exists := videoInfoMap[request.ID]
	if exists {
		videoInfo.CutEditDetails = append(videoInfo.CutEditDetails, CutEditRequest{
			StartTime:  request.StartTime,
			EndTime:    request.EndTime,
			OutputPath: outputFilePath,
		})
		videoInfoMap[request.ID] = videoInfo
	} else {
		videoInfoMap[request.ID] = VideoInfo{
			ID:               request.ID,
			OriginalFilePath: videoFile.FilePath,
			CutEditDetails: []CutEditRequest{
				{
					StartTime:  request.StartTime,
					EndTime:    request.EndTime,
					OutputPath: outputFilePath,
				},
			},
			ConcatDetails:  []ConcatRequest{},
			FinalVideoPath: outputFilePath,
		}
	}

	// 컷 편집 결과 파일 경로 저장
	videoTaskResult.CutFilePaths = append(videoTaskResult.CutFilePaths, outputFilePath)

	// 작업 완료 후 클라이언트에 응답
	c.JSON(http.StatusOK, gin.H{
		"message":     "Cut editing completed successfully",
		"output_file": outputFilePath,
	})
}

// 동영상을 이어 붙이는 함수 (비동기 처리)
func ConcatVideos(c *gin.Context) {
	// 멀티파트 폼 파일을 파싱 (32MB 크기 제한)
	err := c.Request.ParseMultipartForm(32 << 20)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid form data",
			"message": "Failed to parse multipart form, please check the form data.",
			"details": err.Error(),
		})
		return
	}

	files := c.Request.MultipartForm.File["videos"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "No videos provided",
			"message": "You must upload at least one video file.",
		})
		return
	}

	var filePaths []string
	var videoIDs []string // 이어붙일 동영상의 ID 목록

	for _, fileHeader := range files {
		// 파일 열기
		file, err := fileHeader.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to open file",
				"message": fmt.Sprintf("Error opening file: %s", fileHeader.Filename),
				"details": err.Error(),
			})
			return
		}
		defer file.Close()

		fileName := fileHeader.Filename
		filePath := filepath.Join(UploadFolder, fileName)

		// 파일 저장
		out, err := os.Create(filePath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to create file",
				"message": fmt.Sprintf("Error creating file: %s", fileName),
				"details": err.Error(),
			})
			return
		}
		defer out.Close()

		_, err = io.Copy(out, file)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to save file",
				"message": fmt.Sprintf("Error saving file: %s", fileName),
				"details": err.Error(),
			})
			return
		}

		absPath, err := filepath.Abs(filePath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to get absolute path",
				"message": fmt.Sprintf("Error getting absolute path for file: %s", fileName),
				"details": err.Error(),
			})
			return
		}

		filePaths = append(filePaths, absPath)
		videoIDs = append(videoIDs, fileName)
	}

	// 비동기적으로 작업 큐에 FFmpeg 이어붙이기 작업 추가
	go func() {
		concatFilePath := filepath.Join(UploadFolder, "filelist.txt")
		concatFile, err := os.Create(concatFilePath)
		if err != nil {
			fmt.Println("Failed to create filelist.txt:", err)
			return
		}
		defer concatFile.Close()

		for _, filePath := range filePaths {
			_, err := concatFile.WriteString(fmt.Sprintf("file '%s'\n", filepath.ToSlash(filePath)))
			if err != nil {
				fmt.Println("Failed to write to filelist.txt:", err)
				return
			}
		}

		outputFilePath := filepath.Join(UploadFolder, fmt.Sprintf("concat_%d.mp4", time.Now().UnixNano()))

		// FFmpeg 명령어 실행
		cmd := exec.Command("ffmpeg", "-f", "concat", "-safe", "0", "-i", concatFilePath, "-c", "copy", outputFilePath)
		output, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Println("Failed to execute FFmpeg command:", err, string(output))
			return
		}

		// 작업 완료 후 결과 저장
		videoTaskResult.ConcatFilePaths = append(videoTaskResult.ConcatFilePaths, outputFilePath)
		fmt.Println("Concat completed:", outputFilePath)
	}()

	// 클라이언트에게 작업 시작을 알림
	c.JSON(http.StatusOK, gin.H{
		"message":     "Concat task started successfully",
		"video_ids":   videoIDs,
		"concat_task": "Processing in background",
	})
}

func ReExecutePreviousTasks(c *gin.Context) {
	// 컷 편집 작업을 다시 실행 (경로만 가지고 재실행)
	for _, cutFilePath := range videoTaskResult.CutFilePaths {
		// 새로운 파일 경로 생성
		outputFilePath := strings.Replace(cutFilePath, "cut_", "re_cut_", 1)

		// FFmpeg 명령 실행 (기존 컷 편집 파일 사용)
		cmd := exec.Command("ffmpeg", "-i", cutFilePath, "-c", "copy", outputFilePath)
		output, err := cmd.CombinedOutput() // 명령 출력 로그 캡처
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":         "Failed to re-execute cut task",
				"details":       err.Error(),
				"ffmpeg_output": string(output), // 출력 로그 반환
			})
			return
		}
		// 컷 편집된 새로운 파일을 결과 목록에 추가
		videoTaskResult.CutFilePaths = append(videoTaskResult.CutFilePaths, outputFilePath)
	}

	// 이어붙이기 작업을 위한 파일 경로 리스트 생성
	concatFilePath := filepath.Join(UploadFolder, "re_filelist.txt")
	concatFile, err := os.Create(concatFilePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create re_filelist.txt"})
		return
	}
	defer concatFile.Close()

	// 이어붙일 파일들의 절대 경로를 filelist.txt에 작성
	for _, concatFilePath := range videoTaskResult.ConcatFilePaths {
		absPath, err := filepath.Abs(concatFilePath) // 절대 경로 변환
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get absolute path"})
			return
		}

		// 경로를 슬래시("/") 형식으로 변환하여 FFmpeg에 맞게 처리
		_, err = concatFile.WriteString(fmt.Sprintf("file '%s'\n", filepath.ToSlash(absPath)))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write to re_filelist.txt"})
			return
		}
	}

	// 새로 이어붙일 결과 파일 경로 설정
	outputFilePath := filepath.Join(UploadFolder, fmt.Sprintf("re_concat_%d.mp4", time.Now().UnixNano()))

	// FFmpeg 이어붙이기 명령 실행
	cmd := exec.Command("ffmpeg", "-f", "concat", "-safe", "0", "-i", concatFilePath, "-c", "copy", outputFilePath)
	output, err := cmd.CombinedOutput() // 명령 출력 로그 캡처
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":         "Failed to re-execute concat task",
			"details":       err.Error(),
			"ffmpeg_output": string(output), // 출력 로그 반환
		})
		return
	}

	// 이어붙이기 완료 후 결과 파일을 ConcatFilePaths에 추가
	videoTaskResult.ConcatFilePaths = append(videoTaskResult.ConcatFilePaths, outputFilePath)

	c.JSON(http.StatusOK, gin.H{
		"message":     "Previous tasks re-executed successfully",
		"output_file": outputFilePath,
	})
}

// 최종 동영상 다운로드 API (원본, 컷 편집 또는 이어붙이기 동영상 ID를 기반으로)
func DownloadFinalVideo(c *gin.Context) {
	// 요청에서 동영상 ID를 쿼리 파라미터로 받음
	videoID := c.Query("id")

	// ID가 제공되지 않았을 때의 오류 처리
	if videoID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "No video ID provided",
			"message": "Please provide a valid video ID to download the final video.",
		})
		return
	}

	// 동영상 경로를 저장할 변수
	var finalVideoPath string
	var exists bool

	// 원본 동영상 경로 먼저 확인
	if videoInfo, ok := videoInfoMap[videoID]; ok {
		finalVideoPath = videoInfo.OriginalFilePath
		exists = true
	}

	// 컷 편집 파일에서 ID 확인
	if !exists {
		for _, cutFilePath := range videoTaskResult.CutFilePaths {
			if strings.Contains(cutFilePath, videoID) {
				finalVideoPath = cutFilePath
				exists = true
				break
			}
		}
	}

	// 이어붙인 파일에서 ID 확인
	if !exists {
		for _, concatFilePath := range videoTaskResult.ConcatFilePaths {
			if strings.Contains(concatFilePath, videoID) {
				finalVideoPath = concatFilePath
				exists = true
				break
			}
		}
	}

	// 동영상이 존재하지 않는 경우
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Final video not found",
			"message": fmt.Sprintf("No video found for the provided ID: %s", videoID),
		})
		return
	}

	// 동영상 파일이 존재하는지 확인
	if _, err := os.Stat(finalVideoPath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "File does not exist",
			"message": fmt.Sprintf("The file for video ID %s does not exist.", videoID),
		})
		return
	} else if err != nil {
		// 파일 상태를 확인하는 도중 오류가 발생했을 경우
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to check file existence",
			"details": err.Error(),
		})
		return
	}

	// 상대 경로에서 'uploads/'가 이미 포함되어 있는지 확인하고 중복을 방지
	relativePath := strings.TrimPrefix(finalVideoPath, UploadFolder)
	relativePath = strings.TrimPrefix(relativePath, "uploads/")

	// 다운로드 링크 생성
	downloadURL := fmt.Sprintf("http://%s/%s", c.Request.Host, relativePath)

	// videoInfoMap에 최종 동영상 경로 저장
	if videoInfo, exists := videoInfoMap[videoID]; exists {
		// FinalVideoPath 업데이트 (최종 경로만 업데이트)
		videoInfo.FinalVideoPath = finalVideoPath
		videoInfoMap[videoID] = videoInfo
	} else {
		// videoInfoMap에 새로운 정보 추가
		videoInfoMap[videoID] = VideoInfo{
			ID:               videoID,
			OriginalFilePath: finalVideoPath,
			FinalVideoPath:   finalVideoPath,
			CutEditDetails:   []CutEditRequest{}, // 기본값
			ConcatDetails:    []ConcatRequest{},  // 기본값
		}
	}

	// 다운로드 링크 반환
	c.JSON(http.StatusOK, gin.H{
		"message":      "Download link generated successfully",
		"download_url": downloadURL,
	})
}

// 동영상 정보 조회 API
func GetVideoInfo(c *gin.Context) {
	videoID := c.Query("id") // 동영상 ID를 쿼리 파라미터로 받음

	// 동영상 ID가 없으면 전체 목록을 반환
	if videoID == "" {
		// videoInfoMap이 비어 있을 경우 처리
		if len(videoInfoMap) == 0 {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "No video information found",
				"message": "There are no videos uploaded or processed yet.",
			})
			return
		}

		var videoInfos []VideoInfo
		for _, info := range videoInfoMap {
			videoInfos = append(videoInfos, info)
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "All video information retrieved successfully",
			"data":    videoInfos,
		})
		return
	}

	// 특정 동영상 ID에 대한 정보 조회
	videoInfo, exists := videoInfoMap[videoID]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Video not found",
			"message": fmt.Sprintf("No video found for the provided ID: %s", videoID),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Video information retrieved successfully",
		"data":    videoInfo,
	})
}

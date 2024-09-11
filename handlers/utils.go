package handlers

import (
	"fmt"
	"os"
	"os/exec"
)

// FFmpeg 명령어 실행 함수
func executeFFmpegCommands(taskID string, taskCommands []Command) (string, error) {
	outputFile := fmt.Sprintf("%s_output.mp4", taskID)
	concatFile := fmt.Sprintf("%s_concat.txt", taskID)

	for _, cmd := range taskCommands {
		switch cmd.Type {
		case "cut":
			// FFmpeg 컷 편집 명령 실행
			cutCmd := exec.Command("ffmpeg", "-i", cmd.FilePath, "-ss", cmd.StartTime, "-to", cmd.EndTime, "-c", "copy", fmt.Sprintf("cut_%s", cmd.FilePath))
			if err := cutCmd.Run(); err != nil {
				return "", err
			}
		case "concat":
			// 이어 붙이기 명령어 파일 생성
			file, err := os.Create(concatFile)
			if err != nil {
				return "", err
			}
			defer file.Close()

			for _, filePath := range cmd.FilePaths {
				file.WriteString(fmt.Sprintf("file '%s'\n", filePath))
			}

			// FFmpeg 이어 붙이기 명령 실행
			concatCmd := exec.Command("ffmpeg", "-f", "concat", "-safe", "0", "-i", concatFile, "-c", "copy", outputFile)
			if err := concatCmd.Run(); err != nil {
				return "", err
			}
		}
	}

	return outputFile, nil
}

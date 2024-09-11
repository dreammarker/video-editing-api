package handlers

// 명령 구조체
type Command struct {
	Type      string   `json:"type"`                 // 명령 타입 ("cut", "concat")
	FilePath  string   `json:"file_path,omitempty"`  // 컷 편집 시 파일 경로
	StartTime string   `json:"start_time,omitempty"` // 컷 편집 시작 시간
	EndTime   string   `json:"end_time,omitempty"`   // 컷 편집 종료 시간
	FilePaths []string `json:"file_paths,omitempty"` // 이어 붙이기 시 파일 경로 목록
}

# 동영상 편집 API

이 프로젝트는 **Go 언어**를 사용하여 **FFmpeg** 기반의 동영상 편집 API를 제공합니다. 사용자는 동영상을 업로드하고, 컷 편집 및 이어붙이기 작업을 진행한 후 최종 동영상을 다운로드할 수 있습니다.

## 프로젝트 설정

### 필수 설치

1. **Go** (버전 1.16 이상)
   - [Go 다운로드](https://golang.org/dl/)
   
2. **FFmpeg** (시스템 경로에 FFmpeg가 설치되어 있어야 합니다)
   - [FFmpeg 다운로드](https://ffmpeg.org/download.html)
   
3. **VSCode** (또는 선호하는 코드 편집기)
   - [VSCode 다운로드](https://code.visualstudio.com/)

### 환경 설정 (Windows 환경 기준)

1. **FFmpeg 설치**:
   - FFmpeg를 [다운로드 페이지](https://ffmpeg.org/download.html)에서 다운로드하세요.
   - 다운로드한 압축 파일을 해제하고, `bin` 폴더 경로를 시스템 환경 변수(`PATH`)에 추가하세요.
   - FFmpeg가 설치되었는지 확인하려면, **명령 프롬프트(cmd)**를 열고 `ffmpeg -version` 명령을 실행하세요.

2. **Go 설치**:
   - Go를 설치하고, 환경 변수를 설정하여 `go version` 명령으로 정상적으로 설치되었는지 확인하세요.

3. **프로젝트 클론 및 의존성 설치**:

   ```bash
   git clone https://github.com/yourusername/video-editing-api.git
   cd video-editing-api

4. 의존성 설치 
    ```
    go mod tidy
    ```

5. 업로드 디렉터리 생성:
    ```
    mkdir uploads
    ```
    
## 주요 패키지

- Gin: Go에서 HTTP 웹 프레임워크로 사용됩니다.
    ```
    go get github.com/gin-gonic/gin
    ```
- UUID: 고유 ID 생성을 위한 패키지입니다.
    ```
    go get github.com/google/uuid
    ```

## 프로젝트 실행 

    go run main.go


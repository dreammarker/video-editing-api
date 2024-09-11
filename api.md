
---
1.동영상 업로드
- Endpoint: /upload
- Method: POST
- URL: http://localhost:8080/upload
- Description: 하나 이상의 동영상 파일을 업로드 합니다.

Postman 설정
1. HTTP 메서드: POST
2. URL: http://localhost:8080/upload
3. 헤더:
    - Content-Type: multipart/form-data
4. Body:
    - Postman에서 Body 형식을 form-data로 설정하고 videos 필드에 파일을 추가하세요.

Postman 예시
```
POST /upload
Content-Type: multipart/form-data
```
Body 예시
|Key|Type|Value|
|---|---|---|
|videos|File|path/to/video1.mp4|
|videos|File|path/to/video2.mov|

응답 예시
```
{
  "message": "Files uploaded successfully",
  "uploaded_files": [
    {
      "id": "c9f0f75c-532e-402f-9708-26c4247f5891",
      "file_path": "uploads/c9f0f75c-532e-402f-9708-26c4247f5891.mp4"
    },
    {
      "id": "59137557-01f6-45f9-8b29-30284b37beab",
      "file_path": "uploads/59137557-01f6-45f9-8b29-30284b37beab.mov"
    }
  ]
}
```
---
2. 동영상 컷 편집
- Endpoint: /trim
- Method: POST
- URL: http://localhost:8080/trim
- Description: 업로드된 동영상에서 시작 시간과 종료 시간을 사용해 원하는 부분을 잘라냅니다.

Postman 설정
1. HTTP 메서드: POST
2. URL: http://localhost:8080/trim
3. 헤더:
    - Content-Type: application/json
4. Body (raw -JSON):
```
{
  "id": "c9f0f75c-532e-402f-9708-26c4247f5891",
  "start_time": "00:00:10",
  "end_time": "00:00:30"
}
```
응답 예시
```
{
  "message": "Cut editing completed successfully",
  "output_file": "uploads/cut_c9f0f75c-532e-402f-9708-26c4247f5891_00-00-10.mp4"
}
```
---
3. 동영상 이어붙이기 
- Endpoint: /concat
- Method: POST
- URL: http://localhost:8080/concat
- Description: 여러 개의 동영상 파일을 이어붙입니다.


Postman 설정
1. HTTP 메서드: POST
2. URL: http://localhost:8080/concat
3. 헤더:
    - Content-Type: multipart/form-data
4. Body:
Postman에서 Body 형식을 form-data로 설정하고 videos 필드에 파일을 추가하세요

Postman 예시
```
POST /upload
Content-Type: multipart/form-data
```
Body 예시


|Key|	Type|	Value|
|------|---|---|
|videos|File|path/to/video1.mp4|
|videos|File|path/to/video2.mov|

응답 예시
```
{
  "message": "Videos concatenated successfully",
  "output_file": "uploads/concat_1625838264.mp4"
}
```

4. 동영상 정보 조회
- Endpoint: /video-info
- Method: POST
- URL: http://localhost:8080/video-info
- Description: 업로드된 동영상과 편집된 동영상 정보를 조회합니다.


쿼리 파라미터:
|Key|Value|
|------|---|
|id	|c9f0f75c-532e-402f-9708-26c4247f5891|

```
{
  "message": "Video information retrieved successfully",
  "data": {
    "ID": "c9f0f75c-532e-402f-9708-26c4247f5891",
    "OriginalFilePath": "uploads/c9f0f75c-532e-402f-9708-26c4247f5891.mp4",
    "CutEditDetails": [
      {
        "StartTime": "00:00:10",
        "EndTime": "00:00:30",
        "OutputPath": "uploads/cut_c9f0f75c-532e-402f-9708-26c4247f5891_00-00-10.mp4"
      }
    ],
    "ConcatDetails": [],
    "FinalVideoPath": ""
  }
}
```
5. 최종 동영상 다운로드
- Endpoint: /download
- Method: GET
- URL: http://localhost:8080/download
- Description: 최종 처리된 동영상(컷 편집 또는 이어붙인 동영상)의 다운로드 링크를 생성합니다.

Postman 설정
1. HTTP 메서드: GET
2. URL: http://localhost:8080downloadid=c9f0f75c-532e-402f-9708-26c4247f5891

쿼리 파라미터: 
|Key|Value|
|---|---|
|id|c9f0f75c-532e-402f-9708-26c4247f5891|

응답 예시:
```
{
  "message": "Download link generated successfully",
  "download_url": "http://localhost:8080/uploads/cut_c9f0f75c-532e-402f-9708-26c4247f5891_00-00-10.mp4"
}
```
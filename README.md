# 🚀 개요
이 레포지토리는 **NewsClip 서비스의 백엔드 서버**로,  
Go 언어(Gin 프레임워크)를 기반으로 제작된 RESTful API 서버입니다.  

- 📰 네이버 뉴스 API를 활용한 뉴스 수집  
- 🤖 OpenAI API를 통한 뉴스 요약 기능  
- 💬 커뮤니티 / 댓글 / 알림 시스템  
- 🔐 JWT 기반 인증 및 권한 관리  
- 🗄 PostgreSQL 데이터베이스 연동  

프론트엔드(Flutter) 앱과 통신하여 실시간 데이터 제공 및 사용자 활동을 처리합니다.

---

## ⚙️ 개발 환경

| 항목 | 내용 |
|------|------|
| Go Version | **1.24.9** |
| Framework | **Gin Web Framework** |
| Database | **PostgreSQL 15+** |
| ORM | **GORM v2** |
| Auth | **JWT (github.com/golang-jwt/jwt/v5)** |
| API 연동 | **Naver News API**, **OpenAI API** |
| 환경 관리 | `.env` 파일 기반 (godotenv 사용) |
| 배포 환경 | Ubuntu 24.04 LTS (Nginx Reverse Proxy + Certbot SSL) |

---

## 📂 디렉토리 구조

```
backend/
├── cmd/
│   └── main.go                     # 서버 실행 진입점
│
├── config/
│   ├── config.go                   # 환경 변수 로드 및 초기 설정
│   └── database.go                 # PostgreSQL 연결 설정
│
├── internal/
│   ├── app/                        # 핵심 애플리케이션 로직
│   │   ├── controllers/            # HTTP 핸들러 (Gin 컨트롤러)
│   │   │   ├── auth_controller.go
│   │   │   ├── news_controller.go
│   │   │   ├── shorts_controller.go
│   │   │   ├── community_controller.go
│   │   │   ├── comment_controller.go
│   │   │   └── admin_controller.go
│   │   │
│   │   ├── services/               # 비즈니스 로직
│   │   │   ├── auth_service.go
│   │   │   ├── news_service.go
│   │   │   ├── post_service.go
│   │   │   ├── short_service.go
│   │   │   ├── comment_service.go
│   │   │   └── notification_service.go
│   │   │
│   │   ├── repositories/           # 데이터베이스 접근 계층
│   │   │   ├── user_repository.go
│   │   │   ├── news_repository.go
│   │   │   ├── post_repository.go
│   │   │   └── comment_repository.go
│   │   │
│   │   ├── models/                 # DB 모델 구조체 (GORM 기반)
│   │   │   ├── user.go
│   │   │   ├── news.go
│   │   │   ├── post.go
│   │   │   ├── short.go
│   │   │   ├── comment.go
│   │   │   └── report.go
│   │   │
│   │   ├── middlewares/            # 인증 / 로깅 / CORS / 에러핸들링
│   │   │   ├── auth_middleware.go
│   │   │   ├── cors_middleware.go
│   │   │   └── logging_middleware.go
│   │   │
│   │   ├── routes/                 # 라우팅 정의
│   │   │   └── router.go
│   │   │
│   │   └── utils/                  # 공용 유틸리티
│   │       ├── jwt.go
│   │       ├── password.go
│   │       └── response.go
│   │
│   ├── migrations/                 # DB 마이그레이션 SQL 파일
│   │   ├── 001_create_users.sql
│   │   ├── 002_create_news.sql
│   │   └── ...
│   │
│   └── seeds/                      # 초기 데이터 (예: 관리자 계정)
│       └── seed_admin.go
│
├── pkg/                            # 외부 패키지 및 Helper 모듈
│   ├── openai/                     # OpenAI API 요약 모듈
│   └── navernews/                  # 네이버 뉴스 API 클라이언트
│
├── test/                           # 단위 테스트 코드
│   ├── auth_test.go
│   ├── news_test.go
│   └── post_test.go
│
├── .env                            # 환경 변수 (DB_URL, JWT_SECRET 등)
├── go.mod
├── go.sum
└── README.md
```

---

## 🔑 환경 변수 (.env 예시)

```env
# Server
PORT=8080
GIN_MODE=release

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=yourpassword
DB_NAME=newsclip

# JWT
JWT_SECRET=super_secret_key

# External APIs
NAVER_CLIENT_ID=your_client_id
NAVER_CLIENT_SECRET=your_client_secret
OPENAI_API_KEY=your_openai_api_key
```

---

## 🧠 서버 실행 방법

### 1️⃣ 의존성 설치
```bash
go mod tidy
```

### 2️⃣ 환경 변수 설정
```bash
cp .env.example .env
# 내용 수정 후 저장
```

### 3️⃣ 서버 실행
```bash
go run cmd/main.go
```

### 4️⃣ 확인
```bash
curl http://localhost:8080/v1/ping
# {"message":"pong"}
```

---

## 🧩 주요 기능

| 모듈 | 기능 |
|------|------|
| **Auth** | 회원가입, 로그인, JWT 인증 |
| **News** | 네이버 뉴스 API 연동, 뉴스 좋아요/북마크 |
| **Shorts** | OpenAI 요약, 릴스 형식 뉴스 피드 |
| **Community** | 게시글, 댓글 CRUD, 전문가/일반 분리 |
| **Notification** | 키워드 기반 푸시 알림 |
| **Admin** | 신고 처리, 유저 정지, 권한 변경 |

---

## 🧪 테스트 실행

```bash
go test ./test/...
```

---

## 🌐 배포 관련

| 항목 | 내용 |
|------|------|
| 서버 IP | 40.81.180.143 |
| Base URL | https://newsclip.duckdns.org/v1 |
| Reverse Proxy | Nginx + Certbot SSL |
| Database | PostgreSQL (Docker 또는 로컬) |

---

## 🤝 협업 규칙

- **main 브랜치**: 안정화된 배포용 코드  
- **dev 브랜치**: 개발 통합용 (PR 머지 전 테스트 완료 필수)  
- **feature/** 브랜치: 각 기능 단위 (예: `feature/auth`, `feature/news`)  

PR 시에는 반드시 코드 리뷰를 요청합니다.

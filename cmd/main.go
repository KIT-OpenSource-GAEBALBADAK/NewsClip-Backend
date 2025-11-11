package main

import (
	"log"
	"newsclip/backend/config"
	"newsclip/backend/internal/app/models"
	"newsclip/backend/internal/app/routes"
	"newsclip/backend/internal/app/services"

	"github.com/robfig/cron/v3"
)

// DB 마이그레이션 함수
func MigrateDB() {
	err := config.DB.AutoMigrate(
		&models.User{},
		&models.UserSetting{},
		&models.Session{},
		&models.News{},
		// &models.NewsLike{},      // [삭제]
		&models.NewsBookmark{},
		&models.NewsComment{},
		&models.Short{},
		// &models.ShortLike{},     // [삭제]
		&models.ShortComment{},
		&models.Post{},
		&models.PostLike{},
		&models.PostComment{},
		&models.AlertKeyword{},
		&models.Notification{},
		&models.Report{},
		&models.Ban{},

		// === [신규] ===
		&models.UserPreferredCategory{},
		&models.PostImage{},
		&models.NewsInteraction{},
		&models.ShortInteraction{},
	)
	if err != nil {
		log.Fatal("Failed to migrate database")
	}
	log.Println("🚀 Database migration completed!")
}

// === [수정] StartNewsPolling ===
func StartNewsPolling() {
	log.Println("⏰ Starting background news polling...")

	c := cron.New()

	// [수정] "@every 30m" -> "@every 3h" (3시간마다)
	c.AddFunc("@every 3h", func() {
		log.Println("📰 [Cron Job] Starting scheduled fetch for all categories...")

		// FetchAllCategories는 이제 5개씩 가져옵니다.
		err := services.FetchAllCategories()

		if err != nil {
			log.Printf("🔥 [Cron Job] FAILED: %v\n", err)
		} else {
			log.Println("👍 [Cron Job] All categories fetch finished successfully.")
		}
	})

	c.Start()
}

// === [신규] 오래된 뉴스 삭제 스케줄러 ===
func StartCleanupScheduler() {
	log.Println("🧹 Starting old news cleanup scheduler...")
	c := cron.New()

	// "@daily" = 매일 자정 00:00 에 실행
	c.AddFunc("@daily", func() {
		log.Println("🌙 [Cleaner Job] Running daily cleanup for news older than 14 days...")
		// 에러는 서비스 내부에서 로깅
		services.CleanupOldNews()
	})

	c.Start()
}

func main() {
	// 1. 환경 변수 로드
	config.LoadConfig()

	// 2. 데이터베이스 연결
	config.ConnectDB()

	// 3. 데이터베이스 마이그레이션 실행
	MigrateDB()

	// 4. === [신규] 스케줄러 시작 ===
	// 4.1. (수정) go StartNewsPolling()
	//    스케줄러는 백그라운드에서 실행
	go StartNewsPolling()

	// [신규] 뉴스 삭제 스케줄러 시작
	go StartCleanupScheduler()

	// 4.2. (추가) 서버 시작 시 1회 즉시 실행
	log.Println("🚀 Running initial poll ONCE for all categories...")
	err := services.FetchAllCategories()
	if err != nil {
		log.Printf("🔥 INITIAL POLL FAILED: %v\n", err)
	} else {
		log.Println("👍 INITIAL POLL SUCCEEDED.")
	}
	// 5. 라우터 설정
	router := routes.SetupRouter()

	// 6. 서버 실행
	router.Run(":8080")
}

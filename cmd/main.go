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

// === StartNewsPolling (스케줄러) ===
func StartNewsPolling() {
	log.Println("⏰ Starting background news polling...")

	c := cron.New()

	// 3시간마다 실행
	c.AddFunc("@every 3h", func() {
		log.Println("📰 [Cron Job] 1. Fetching News...")

		// 1. 뉴스 수집
		err := services.FetchAllCategories()
		if err != nil {
			log.Printf("🔥 News Fetch Failed: %v", err)
			return // 뉴스 수집 실패하면 쇼츠 생성도 중단
		}

		// 2. 쇼츠 생성 (뉴스 수집 완료 후 실행)
		log.Println("🤖 [Cron Job] 2. Generating Shorts...")
		err = services.GenerateShorts()
		if err != nil {
			log.Printf("🔥 Shorts Generation Failed: %v", err)
		}
	})

	c.Start()
}

// === 오래된 뉴스 삭제 스케줄러 ===
func StartCleanupScheduler() {
	log.Println("🧹 Starting old news cleanup scheduler...")
	c := cron.New()

	// "@daily" = 매일 자정 00:00 에 실행
	c.AddFunc("@daily", func() {
		log.Println("🌙 [Cleaner Job] Running daily cleanup for news older than 14 days...")
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

	// 4. 스케줄러 시작 (백그라운드)
	go StartNewsPolling()
	go StartCleanupScheduler()

	// ==========================================
	// 5. [테스트용] 서버 시작 시 즉시 1회 실행 로직
	// ==========================================
	log.Println("🚀 [TEST MODE] Running initial logic ONCE...")

	// 5-1. 뉴스 수집 실행
	log.Println("📰 1. Fetching News immediately...")
	err := services.FetchAllCategories()
	if err != nil {
		log.Printf("🔥 INITIAL POLL FAILED: %v\n", err)
	} else {
		log.Println("✅ INITIAL POLL SUCCEEDED.")

		// 5-2. 쇼츠 생성 실행 (뉴스 수집 성공 시에만 실행)
		log.Println("🤖 2. Generating Shorts immediately...")
		err = services.GenerateShorts()
		if err != nil {
			log.Printf("🔥 INITIAL SHORTS GENERATION FAILED: %v\n", err)
		} else {
			log.Println("✅ INITIAL SHORTS GENERATION SUCCEEDED.")
		}
	}
	// ==========================================

	// 6. 라우터 설정 및 서버 실행
	router := routes.SetupRouter()
	router.Run(":8080")
}

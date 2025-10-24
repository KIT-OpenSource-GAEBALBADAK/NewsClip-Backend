package main

import (
	"log"
	"newsclip/backend/config"
	"newsclip/backend/internal/app/models"
	"newsclip/backend/internal/app/routes"
)

// DB 마이그레이션 함수
func MigrateDB() {
	err := config.DB.AutoMigrate(
		&models.User{},
		&models.UserSetting{},
		&models.Session{},
		&models.News{},
		&models.NewsLike{},
		&models.NewsBookmark{},
		&models.NewsComment{},
		&models.Short{},
		&models.ShortLike{},
		&models.ShortComment{},
		&models.Post{},
		&models.PostLike{},
		&models.PostComment{},
		&models.AlertKeyword{},
		&models.Notification{},
		&models.Report{},
		&models.Ban{},
	)
	if err != nil {
		log.Fatal("Failed to migrate database")
	}
	log.Println("🚀 Database migration completed!")
}

func main() {
	// 1. 환경 변수 로드
	config.LoadConfig()

	// 2. 데이터베이스 연결
	config.ConnectDB()

	// 3. 데이터베이스 마이그레이션 실행
	MigrateDB()

	// 4. 라우터 설정
	router := routes.SetupRouter()

	// 5. 서버 실행
	router.Run(":8080")
}

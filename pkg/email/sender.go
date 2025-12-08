package email

import (
	"fmt"
	"newsclip/backend/config"
	"strconv"

	"gopkg.in/gomail.v2"
)

// 인증번호 이메일 발송
func SendVerificationCode(toEmail string, code string) error {
	smtpHost := config.GetEnv("SMTP_HOST")
	smtpPortStr := config.GetEnv("SMTP_PORT")
	smtpEmail := config.GetEnv("SMTP_EMAIL")
	smtpPassword := config.GetEnv("SMTP_PASSWORD")

	smtpPort, _ := strconv.Atoi(smtpPortStr)

	m := gomail.NewMessage()

	m.SetHeader("From", m.FormatAddress(smtpEmail, "NewsClip"))

	m.SetHeader("To", toEmail)
	m.SetHeader("Subject", "[NewsClip] 인증번호가 도착했습니다.")

	// HTML 형식의 이메일 본문
	body := fmt.Sprintf(`
		<div style="font-family: Arial, sans-serif; padding: 20px;">
			<h2>NewsClip 이메일 인증</h2>
			<p>아래 인증번호 6자리를 입력하여 인증을 완료해주세요.</p>
			<h1 style="color: #4CAF50; letter-spacing: 5px;">%s</h1>
			<p>이 코드는 3분간 유효합니다.</p>
		</div>
	`, code)

	m.SetBody("text/html", body)

	d := gomail.NewDialer(smtpHost, smtpPort, smtpEmail, smtpPassword)

	// 메일 전송
	if err := d.DialAndSend(m); err != nil {
		return err
	}
	return nil
}

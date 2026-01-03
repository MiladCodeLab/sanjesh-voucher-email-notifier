package main

import (
	"crypto/tls"
	"fmt"
	"github.com/joho/godotenv"
	gomail "gopkg.in/mail.v2"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("failed to load .env:", err)
	}

	checkURL := mustEnv("CHECK_URL")
	notifyEmail := mustEnv("NOTIFY_EMAIL")
	intervalHours := envInt("CHECK_INTERVAL_HOURS", 1)
	interval := time.Duration(intervalHours) * time.Hour

	log.Println("TOEFL iBT watcher started")

	for {
		qty, err := fetchTOEFLQty(checkURL)
		if err != nil {
			log.Println("check error:", err)
		} else {
			log.Println("TOEFL iBT Qty:", qty)

			if qty > 0 {
				err = sendMail(
					notifyEmail,
					"TOEFL iBT AVAILABLE",
					fmt.Sprintf(
						"Good news!\n\nTOEFL iBT capacity is now %d.\n\n%s",
						qty,
						checkURL,
					),
				)
				if err != nil {
					log.Println("email error:", err)
				} else {
					log.Println("notification email sent")
				}
			}
		}

		time.Sleep(interval)
	}
}

func fetchTOEFLQty(url string) (int, error) {
	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	re := regexp.MustCompile(
		`TOEFL\s*\(iBT\)[\s\S]*?<input[^>]+id="Qty"[^>]+value='(\d+)'`,
	)

	m := re.FindStringSubmatch(string(body))
	if len(m) < 2 {
		return 0, fmt.Errorf("TOEFL quantity not found")
	}

	return strconv.Atoi(m[1])
}

func sendMail(to, subject, body string) error {
	host := mustEnv("SMTP_HOST")
	from := mustEnv("SMTP_FROM")
	password := mustEnv("SMTP_PASSWORD")
	port := envInt("SMTP_PORT", 587)
	user := mustEnv("SMTP_USER")
	message := gomail.NewMessage()
	// Set email headers
	message.SetHeader("From", from)
	message.SetHeader("To", to)
	message.SetHeader("Subject", subject)

	// Set email body
	message.SetBody("text/plain", body)

	// Set up the SMTP dialer
	dialer := gomail.NewDialer(host, port, from, password)
	dialer.StartTLSPolicy = gomail.MandatoryStartTLS

	dialer.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	dialer.Username = user
	// Send the email
	if err := dialer.DialAndSend(message); err != nil {
		fmt.Println("Error:", err)
		return fmt.Errorf("sending mail: %w", err)
	}

	fmt.Println("Email sent successfully!")
	return nil
}

func mustEnv(k string) string {
	v := os.Getenv(k)
	if v == "" {
		log.Fatalf("missing env: %s", k)
	}
	return v
}

func envInt(k string, def int) int {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return i
}

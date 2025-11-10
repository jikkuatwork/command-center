package database

import (
	"fmt"
	"log"
	"math/rand"
	"time"
)

// GenerateMockData inserts sample data for testing
func GenerateMockData() error {
	log.Println("Generating mock data...")

	// Sample domains
	domains := []string{
		"example.com",
		"blog.example.com",
		"shop.example.com",
		"api.example.com",
		"newsletter.com",
	}

	// Sample tags
	tagSets := []string{
		"app,production",
		"blog,content",
		"shop,ecommerce",
		"api,backend",
		"newsletter,email",
		"campaign-123",
		"promo,sale",
		"organic",
		"paid,ads",
		"social,twitter",
	}

	// Sample paths
	paths := []string{
		"/",
		"/about",
		"/contact",
		"/blog",
		"/blog/post-1",
		"/blog/post-2",
		"/products",
		"/products/item-1",
		"/checkout",
		"/api/v1/users",
	}

	// Sample referrers
	referrers := []string{
		"https://google.com",
		"https://twitter.com",
		"https://facebook.com",
		"https://reddit.com",
		"https://newsletter.com",
		"direct",
		"",
	}

	// Sample user agents
	userAgents := []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X)",
		"Mozilla/5.0 (Linux; Android 10) AppleWebKit/537.36",
	}

	// Generate 100 events
	for i := 0; i < 100; i++ {
		domain := domains[rand.Intn(len(domains))]
		tags := tagSets[rand.Intn(len(tagSets))]
		sourceType := []string{"web", "pixel", "redirect"}[rand.Intn(3)]
		eventType := []string{"pageview", "click", "redirect"}[rand.Intn(3)]
		path := paths[rand.Intn(len(paths))]
		referrer := referrers[rand.Intn(len(referrers))]
		userAgent := userAgents[rand.Intn(len(userAgents))]
		ipAddress := fmt.Sprintf("192.168.1.%d", rand.Intn(255))

		// Random timestamp within last 7 days
		hoursAgo := rand.Intn(168) // 7 days * 24 hours
		createdAt := time.Now().Add(-time.Duration(hoursAgo) * time.Hour)

		_, err := db.Exec(`
			INSERT INTO events (domain, tags, source_type, event_type, path, referrer, user_agent, ip_address, created_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, domain, tags, sourceType, eventType, path, referrer, userAgent, ipAddress, createdAt)

		if err != nil {
			return fmt.Errorf("failed to insert event: %w", err)
		}
	}

	// Generate 10 redirects
	redirectSlugs := []string{
		"github",
		"twitter",
		"docs",
		"promo-1",
		"promo-2",
		"newsletter",
		"blog-post",
		"product-launch",
		"sale",
		"demo",
	}

	destinations := []string{
		"https://github.com/example",
		"https://twitter.com/example",
		"https://docs.example.com",
		"https://example.com/promo",
		"https://example.com/special-offer",
		"https://newsletter.example.com/subscribe",
		"https://blog.example.com/post",
		"https://example.com/launch",
		"https://shop.example.com/sale",
		"https://demo.example.com",
	}

	for i := 0; i < 10; i++ {
		slug := redirectSlugs[i]
		destination := destinations[i]
		tags := tagSets[rand.Intn(len(tagSets))]
		clickCount := rand.Intn(100)

		_, err := db.Exec(`
			INSERT INTO redirects (slug, destination, tags, click_count)
			VALUES (?, ?, ?, ?)
		`, slug, destination, tags, clickCount)

		if err != nil {
			return fmt.Errorf("failed to insert redirect: %w", err)
		}
	}

	// Generate 5 webhooks
	webhookNames := []string{
		"GitHub Deployments",
		"Stripe Payments",
		"Custom Integration",
		"CI/CD Pipeline",
		"Monitoring Alerts",
	}

	webhookEndpoints := []string{
		"github-deploy",
		"stripe-events",
		"custom-hook",
		"cicd-status",
		"monitoring",
	}

	for i := 0; i < 5; i++ {
		name := webhookNames[i]
		endpoint := webhookEndpoints[i]
		// First 3 webhooks without secret for easy testing, last 2 with secret
		secret := ""
		if i >= 3 {
			secret = fmt.Sprintf("secret_%d_%d", i, rand.Intn(10000))
		}
		isActive := true

		_, err := db.Exec(`
			INSERT INTO webhooks (name, endpoint, secret, is_active)
			VALUES (?, ?, ?, ?)
		`, name, endpoint, secret, isActive)

		if err != nil {
			return fmt.Errorf("failed to insert webhook: %w", err)
		}
	}

	log.Println("Mock data generated successfully")
	log.Println("- 100 events")
	log.Println("- 10 redirects")
	log.Println("- 5 webhooks")

	return nil
}

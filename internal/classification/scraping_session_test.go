package classification

import (
	"net/http"
	"os"
	"testing"
	"time"
)

func TestScrapingSessionManager(t *testing.T) {
	t.Run("creates_session", func(t *testing.T) {
		ssm := NewScrapingSessionManager()
		session, err := ssm.GetOrCreateSession("example.com")
		
		if err != nil {
			t.Fatalf("Expected no error creating session, got %v", err)
		}
		if session == nil {
			t.Fatal("Expected session to be created")
		}
		if session.GetDomain() != "example.com" {
			t.Errorf("Expected domain to be example.com, got %s", session.GetDomain())
		}
	})

	t.Run("reuses_existing_session", func(t *testing.T) {
		ssm := NewScrapingSessionManager()
		session1, _ := ssm.GetOrCreateSession("example.com")
		session2, _ := ssm.GetOrCreateSession("example.com")

		if session1 != session2 {
			t.Error("Expected same session to be reused")
		}
	})

	t.Run("manages_cookies", func(t *testing.T) {
		ssm := NewScrapingSessionManager()
		session, _ := ssm.GetOrCreateSession("example.com")

		cookie := &http.Cookie{
			Name:  "test",
			Value: "value",
		}
		session.SetCookiesForURL("https://example.com", []*http.Cookie{cookie})

		cookies := session.GetCookiesForURL("https://example.com")
		if len(cookies) == 0 {
			t.Error("Expected cookie to be stored")
		}
		if cookies[0].Name != "test" || cookies[0].Value != "value" {
			t.Errorf("Expected cookie name=test value=value, got name=%s value=%s", cookies[0].Name, cookies[0].Value)
		}
	})

	t.Run("tracks_referer", func(t *testing.T) {
		ssm := NewScrapingSessionManager()
		ssm.GetOrCreateSession("example.com")
		ssm.UpdateReferer("example.com", "https://example.com/previous")

		referer := ssm.GetReferer("example.com")
		if referer != "https://example.com/previous" {
			t.Errorf("Expected referer to be https://example.com/previous, got %s", referer)
		}
	})

	t.Run("respects_disabled_setting", func(t *testing.T) {
		os.Setenv("SCRAPING_SESSION_MANAGEMENT_ENABLED", "false")
		defer os.Unsetenv("SCRAPING_SESSION_MANAGEMENT_ENABLED")

		ssm := NewScrapingSessionManager()
		session, err := ssm.GetOrCreateSession("example.com")
		
		if err != nil {
			t.Fatalf("Expected no error even when disabled, got %v", err)
		}
		// Should still create a temporary session
		if session == nil {
			t.Fatal("Expected temporary session to be created")
		}
	})
}

func TestScrapingSessionManager_IsEnabled(t *testing.T) {
	t.Run("default_enabled", func(t *testing.T) {
		os.Unsetenv("SCRAPING_SESSION_MANAGEMENT_ENABLED")
		ssm := NewScrapingSessionManager()
		if !ssm.IsEnabled() {
			t.Error("Expected session management to be enabled by default")
		}
	})

	t.Run("explicitly_disabled", func(t *testing.T) {
		os.Setenv("SCRAPING_SESSION_MANAGEMENT_ENABLED", "false")
		defer os.Unsetenv("SCRAPING_SESSION_MANAGEMENT_ENABLED")

		ssm := NewScrapingSessionManager()
		if ssm.IsEnabled() {
			t.Error("Expected session management to be disabled")
		}
	})
}

func TestCreateHTTPClientWithSession(t *testing.T) {
	ssm := NewScrapingSessionManager()
	session, _ := ssm.GetOrCreateSession("example.com")

	client := CreateHTTPClientWithSession(session, 30*time.Second)
	if client == nil {
		t.Fatal("Expected HTTP client to be created")
	}
	if client.Jar == nil {
		t.Error("Expected cookie jar to be set")
	}
}


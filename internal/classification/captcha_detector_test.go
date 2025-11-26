package classification

import (
	"net/http/httptest"
	"os"
	"testing"
)

func TestDetectCAPTCHA(t *testing.T) {
	t.Run("detects_recaptcha", func(t *testing.T) {
		detector := NewCAPTCHADetector()
		
		body := []byte(`<html><body><div class="g-recaptcha"></div></body></html>`)
		resp := httptest.NewRecorder()
		resp.Write(body)
		httpResp := resp.Result()

		result := detector.DetectCAPTCHA(httpResp, body)
		if !result.Detected {
			t.Error("Expected reCAPTCHA to be detected")
		}
		if result.Type != CAPTCHATypeReCAPTCHA {
			t.Errorf("Expected CAPTCHA type to be %s, got %s", CAPTCHATypeReCAPTCHA, result.Type)
		}
	})

	t.Run("detects_hcaptcha", func(t *testing.T) {
		detector := NewCAPTCHADetector()
		
		body := []byte(`<html><body><div class="hcaptcha"></div></body></html>`)
		httptest.NewRequest("GET", "/", nil)
		resp := httptest.NewRecorder()
		resp.Write(body)
		httpResp := resp.Result()

		result := detector.DetectCAPTCHA(httpResp, body)
		if !result.Detected {
			t.Error("Expected hCaptcha to be detected")
		}
		if result.Type != CAPTCHATypeHCaptcha {
			t.Errorf("Expected CAPTCHA type to be %s, got %s", CAPTCHATypeHCaptcha, result.Type)
		}
	})

	t.Run("detects_cloudflare", func(t *testing.T) {
		detector := NewCAPTCHADetector()
		
		body := []byte(`<html><body>Checking your browser before accessing...</body></html>`)
		httptest.NewRequest("GET", "/", nil)
		resp := httptest.NewRecorder()
		resp.Header().Set("cf-ray", "123456")
		resp.Write(body)
		httpResp := resp.Result()

		result := detector.DetectCAPTCHA(httpResp, body)
		if !result.Detected {
			t.Error("Expected Cloudflare challenge to be detected")
		}
		if result.Type != CAPTCHATypeCloudflare {
			t.Errorf("Expected CAPTCHA type to be %s, got %s", CAPTCHATypeCloudflare, result.Type)
		}
	})

	t.Run("no_captcha", func(t *testing.T) {
		detector := NewCAPTCHADetector()
		
		body := []byte(`<html><body>Normal page content</body></html>`)
		httptest.NewRequest("GET", "/", nil)
		resp := httptest.NewRecorder()
		resp.Write(body)
		httpResp := resp.Result()

		result := detector.DetectCAPTCHA(httpResp, body)
		if result.Detected {
			t.Error("Expected no CAPTCHA to be detected")
		}
	})

	t.Run("respects_disabled_setting", func(t *testing.T) {
		os.Setenv("SCRAPING_CAPTCHA_DETECTION_ENABLED", "false")
		defer os.Unsetenv("SCRAPING_CAPTCHA_DETECTION_ENABLED")

		detector := NewCAPTCHADetector()
		
		body := []byte(`<html><body><div class="g-recaptcha"></div></body></html>`)
		resp := httptest.NewRecorder()
		resp.Write(body)
		httpResp := resp.Result()

		result := detector.DetectCAPTCHA(httpResp, body)
		if result.Detected {
			t.Error("Expected CAPTCHA detection to be disabled")
		}
	})
}

func TestCAPTCHADetector_IsEnabled(t *testing.T) {
	t.Run("default_enabled", func(t *testing.T) {
		os.Unsetenv("SCRAPING_CAPTCHA_DETECTION_ENABLED")
		detector := NewCAPTCHADetector()
		if !detector.IsEnabled() {
			t.Error("Expected CAPTCHA detection to be enabled by default")
		}
	})

	t.Run("explicitly_disabled", func(t *testing.T) {
		os.Setenv("SCRAPING_CAPTCHA_DETECTION_ENABLED", "false")
		defer os.Unsetenv("SCRAPING_CAPTCHA_DETECTION_ENABLED")

		detector := NewCAPTCHADetector()
		if detector.IsEnabled() {
			t.Error("Expected CAPTCHA detection to be disabled")
		}
	})
}


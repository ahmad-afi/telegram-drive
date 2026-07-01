package middleware

import (
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestSession(t *testing.T) {
	app := fiber.New()
	app.Use(Session())
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString(GetSession(c))
	})

	// missing header → 401
	_, status := get(t, app, "/test", "")
	if status != 401 {
		t.Fatalf("expected 401 for missing session, got %d", status)
	}

	// valid header → 200 with session echoed
	body, status := get(t, app, "/test", "my-session-string")
	if status != 200 {
		t.Fatalf("expected 200, got %d", status)
	}
	if body != "my-session-string" {
		t.Errorf("expected session echoed, got %q", body)
	}

	// whitespace trimmed
	body, _ = get(t, app, "/test", "  spaced-session  ")
	if body != "spaced-session" {
		t.Errorf("expected trimmed session, got %q", body)
	}
}

func get(t *testing.T, app *fiber.App, path, session string) (string, int) {
	t.Helper()
	req := httptest.NewRequest("GET", path, nil)
	if session != "" {
		req.Header.Set(HeaderSession, session)
	}
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(resp.Body)
	return string(body), resp.StatusCode
}

package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

const HeaderSession = "X-Session"

// Session extracts the Telegram session string from the X-Session header
// and stores it in c.Locals. The server stores nothing — the browser owns it.
func Session() fiber.Handler {
	return func(c *fiber.Ctx) error {
		h := c.Get(HeaderSession)
		h = strings.TrimSpace(h)
		if h == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "missing session")
		}
		c.Locals("session", h)
		return c.Next()
	}
}

func GetSession(c *fiber.Ctx) string {
	if v, ok := c.Locals("session").(string); ok {
		return v
	}
	return ""
}

package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/gofiber/fiber/v2"

	"telegram-drive/internal/telegram"
)

type QRHandler struct {
	mgr *telegram.QRManager
}

func NewQRHandler(mgr *telegram.QRManager) *QRHandler {
	return &QRHandler{mgr: mgr}
}

// Start begins a QR login and returns the QR image as PNG base64.
func (h *QRHandler) Start(c *fiber.Ctx) error {
	loginID := randomID()
	if err := h.mgr.Start(loginID); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	img, err := h.mgr.WaitForQR(loginID, 15*time.Second)
	if err != nil {
		return fiber.NewError(fiber.StatusGatewayTimeout, "QR generation: "+err.Error())
	}
	return c.JSON(fiber.Map{
		"loginId": loginID,
		"qrImage": img,
	})
}

// Poll checks if the QR was scanned and accepted.
func (h *QRHandler) Poll(c *fiber.Ctx) error {
	id := c.Params("id")
	sess, name, ok := h.mgr.Poll(id)
	if !ok {
		return c.JSON(fiber.Map{"status": "pending"})
	}
	return c.JSON(fiber.Map{
		"status":   "ok",
		"session":  sess,
		"name":     name,
	})
}

func randomID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

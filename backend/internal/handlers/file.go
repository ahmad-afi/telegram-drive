package handlers

import (
	"bufio"
	"context"
	"time"

	"github.com/gofiber/fiber/v2"

	"telegram-drive/internal/middleware"
	"telegram-drive/internal/telegram"
)

type FileHandler struct{ tg *telegram.Client }

func NewFileHandler(tg *telegram.Client) *FileHandler { return &FileHandler{tg: tg} }

// List returns files in a folder (or root), or searches by name.
// GET /files?folder=root  |  /files?q=term  |  /files?folder=Docs&offset=0&limit=50
func (h *FileHandler) List(c *fiber.Ctx) error {
	sess := middleware.GetSession(c)
	folder := c.Query("folder", "root")
	q := c.Query("q")
	offset := c.QueryInt("offset", 0)
	limit := c.QueryInt("limit", 0)

	ctx, cancel := context.WithTimeout(c.Context(), 30*time.Second)
	defer cancel()

	items, err := h.tg.ListFiles(ctx, sess, folder, q, offset, limit)
	if err != nil {
		return fiber.NewError(fiber.StatusBadGateway, err.Error())
	}
	return c.JSON(items)
}

// Upload sends a file to Saved Messages with a folder caption tag.
func (h *FileHandler) Upload(c *fiber.Ctx) error {
	sess := middleware.GetSession(c)
	fh, err := c.FormFile("file")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "file is required")
	}
	src, err := fh.Open()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	defer src.Close()

	folder := c.FormValue("folder", "root")

	// ponytail: blocking upload tied to request lifetime. Move to a job
	// queue with progress reporting when files routinely exceed ~1GB.
	ctx, cancel := context.WithTimeout(c.Context(), 30*time.Minute)
	defer cancel()

	_, err = h.tg.Upload(ctx, sess, fh.Filename, folder, fh.Size, src)
	if err != nil {
		return fiber.NewError(fiber.StatusBadGateway, "upload failed: "+err.Error())
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"ok": true})
}

// Download streams a file from Telegram to the browser.
func (h *FileHandler) Download(c *fiber.Ctx) error {
	sess := middleware.GetSession(c)
	msgID, err := c.ParamsInt("msgId")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid msg id")
	}
	ctx, cancel := context.WithTimeout(c.Context(), 30*time.Minute)
	defer cancel()

	c.Set("Content-Disposition", `attachment`)
	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		_ = h.tg.Download(ctx, sess, msgID, w)
		_ = w.Flush()
	})
	return nil
}

// Delete removes a file from Telegram.
func (h *FileHandler) Delete(c *fiber.Ctx) error {
	sess := middleware.GetSession(c)
	msgID, err := c.ParamsInt("msgId")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid msg id")
	}
	ctx, cancel := context.WithTimeout(c.Context(), 2*time.Minute)
	defer cancel()

	if err := h.tg.Delete(ctx, sess, msgID); err != nil {
		return fiber.NewError(fiber.StatusBadGateway, err.Error())
	}
	return c.SendStatus(fiber.StatusNoContent)
}

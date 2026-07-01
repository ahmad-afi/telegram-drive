package handlers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"

	"telegram-drive/internal/middleware"
	"telegram-drive/internal/telegram"
)

type FolderHandler struct{ tg *telegram.Client }

func NewFolderHandler(tg *telegram.Client) *FolderHandler { return &FolderHandler{tg: tg} }

// List returns unique folder names extracted from Saved Messages captions.
func (h *FolderHandler) List(c *fiber.Ctx) error {
	sess := middleware.GetSession(c)
	ctx, cancel := context.WithTimeout(c.Context(), 30*time.Second)
	defer cancel()

	folders, err := h.tg.ListFolders(ctx, sess)
	if err != nil {
		return fiber.NewError(fiber.StatusBadGateway, err.Error())
	}
	return c.JSON(folders)
}

// Create is a no-op: folders are created implicitly when a file is uploaded
// with a folder tag. We keep this endpoint for API symmetry.
// ponytail: if you ever want pre-created empty folders, store a placeholder
// message with just the tag caption.
func (h *FolderHandler) Create(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"message": "folders are created automatically when you upload a file with a folder",
	})
}

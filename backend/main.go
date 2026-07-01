package main

import (
	"log"

	"telegram-drive/internal/config"
	"telegram-drive/internal/handlers"
	"telegram-drive/internal/middleware"
	"telegram-drive/internal/telegram"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {
	cfg := config.Load()

	tgClient := telegram.New(cfg.AppID, cfg.AppHash)
	qrMgr := telegram.NewQRManager(cfg.AppID, cfg.AppHash)

	app := fiber.New(fiber.Config{
		BodyLimit: 4 * 1024 * 1024 * 1024, // 4GB
	})
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: cfg.CORS,
		AllowHeaders: "Origin, Content-Type, Accept, X-Session",
		AllowMethods: "GET,POST,DELETE,OPTIONS",
	}))

	// QR login — no session needed
	qrH := handlers.NewQRHandler(qrMgr)
	app.Post("/api/qr/start", qrH.Start)
	app.Get("/api/qr/poll/:id", qrH.Poll)

	// All below routes require X-Session header
	api := app.Group("/api", middleware.Session())

	fileH := handlers.NewFileHandler(tgClient)
	api.Get("/files", fileH.List)
	api.Post("/files", fileH.Upload)
	api.Get("/files/:msgId/download", fileH.Download)
	api.Delete("/files/:msgId", fileH.Delete)

	folderH := handlers.NewFolderHandler(tgClient)
	api.Get("/folders", folderH.List)
	api.Post("/folders", folderH.Create)

	log.Fatal(app.Listen(":" + cfg.Port))
}

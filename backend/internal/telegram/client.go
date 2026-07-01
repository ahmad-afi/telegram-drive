package telegram

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"

	"github.com/gotd/td/session"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/downloader"
	"github.com/gotd/td/telegram/uploader"
	"github.com/gotd/td/tg"
)

type Client struct {
	appID   int
	appHash string
}

func New(appID int, appHash string) *Client {
	return &Client{appID: appID, appHash: appHash}
}

// Run connects to Telegram using the given session string (base64-encoded
// gotd session JSON), executes fn, then discards the connection. The server
// stores nothing.
func (c *Client) Run(ctx context.Context, sessionB64 string, fn func(context.Context, *tg.Client) error) error {
	data, err := decodeSession(sessionB64)
	if err != nil {
		return fmt.Errorf("invalid session: %w", err)
	}
	mem := &session.StorageMemory{}
	if err := mem.StoreSession(ctx, data); err != nil {
		return err
	}
	client := telegram.NewClient(c.appID, c.appHash, telegram.Options{
		SessionStorage: mem,
	})
	return client.Run(ctx, func(ctx context.Context) error {
		st, err := client.Auth().Status(ctx)
		if err != nil {
			return err
		}
		if !st.Authorized {
			return fmt.Errorf("session not authorized")
		}
		return fn(ctx, client.API())
	})
}

func decodeSession(b64 string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(b64)
}

// EncodeSession converts raw gotd session Data into a base64 string for
// transport to the browser.
func EncodeSession(d *session.Data) (string, error) {
	v := struct {
		Version int           `json:"Version"`
		Data    session.Data `json:"Data"`
	}{Version: 1, Data: *d}
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}

// Upload sends a file to Saved Messages with a caption tag marking its folder.
// Returns the new message ID.
func (c *Client) Upload(ctx context.Context, sessionB64, name, folder string, sz int64, r io.Reader) (int, error) {
	var msgID int
	caption := tagCaption(folder, name)
	err := c.Run(ctx, sessionB64, func(ctx context.Context, api *tg.Client) error {
		up := uploader.NewUploader(api)
		f, err := up.Upload(ctx, uploader.NewUpload(name, r, sz))
		if err != nil {
			return err
		}
		media := &tg.InputMediaUploadedDocument{
			File:     f,
			MimeType: "application/octet-stream",
			Attributes: []tg.DocumentAttributeClass{
				&tg.DocumentAttributeFilename{FileName: name},
			},
		}
		updates, err := api.MessagesSendMedia(ctx, &tg.MessagesSendMediaRequest{
			Peer:     &tg.InputPeerSelf{},
			Media:    media,
			Message:  caption,
			RandomID: randID(),
		})
		if err != nil {
			return err
		}
		msgID = extractMsgID(updates)
		if msgID == 0 {
			return fmt.Errorf("could not determine uploaded message id")
		}
		return nil
	})
	return msgID, err
}

// Download streams the document at msgID to w.
func (c *Client) Download(ctx context.Context, sessionB64 string, msgID int, w io.Writer) error {
	return c.Run(ctx, sessionB64, func(ctx context.Context, api *tg.Client) error {
		doc, err := docFromMessage(ctx, api, msgID)
		if err != nil {
			return err
		}
		loc := doc.AsInputDocumentFileLocation("")
		d := downloader.NewDownloader()
		_, err = d.Download(api, loc).Stream(ctx, w)
		return err
	})
}

// Delete removes the message holding the file.
func (c *Client) Delete(ctx context.Context, sessionB64 string, msgID int) error {
	return c.Run(ctx, sessionB64, func(ctx context.Context, api *tg.Client) error {
		_, err := api.MessagesDeleteMessages(ctx, &tg.MessagesDeleteMessagesRequest{
			Revoke: true,
			ID:     []int{msgID},
		})
		return err
	})
}

// FileItem is a parsed file entry from Saved Messages.
type FileItem struct {
	MsgID    int    `json:"msgId"`
	Name     string `json:"name"`
	Size     int64  `json:"size"`
	MimeType string `json:"mimeType"`
	Folder   string `json:"folder"`
	Date     int64  `json:"date"`
}

// ListFiles scans Saved Messages history and returns file entries.
// If folder is "root" or "", returns only files without a folder tag.
// If folder is non-empty, returns files tagged with that folder.
// If q is non-empty, searches file names instead (filtered by folder context).
// offset and limit control pagination (offset=0, limit=0 means no pagination).
func (c *Client) ListFiles(ctx context.Context, sessionB64, folder, q string, offset, limit int) ([]FileItem, error) {
	var items []FileItem
	err := c.Run(ctx, sessionB64, func(ctx context.Context, api *tg.Client) error {
		if q != "" {
			return searchFiles(ctx, api, q, folder, &items)
		}
		return scanHistory(ctx, api, folder, &items)
	})
	if err != nil {
		return nil, err
	}
	if limit > 0 && offset >= 0 {
		if offset >= len(items) {
			return []FileItem{}, nil
		}
		end := offset + limit
		if end > len(items) {
			end = len(items)
		}
		return items[offset:end], nil
	}
	return items, nil
}

// ListFolders extracts unique folder tags from Saved Messages captions.
func (c *Client) ListFolders(ctx context.Context, sessionB64 string) ([]string, error) {
	all, err := c.ListFiles(ctx, sessionB64, "__all__", "", 0, 0)
	if err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	var folders []string
	for _, f := range all {
		if f.Folder != "" && !seen[f.Folder] {
			seen[f.Folder] = true
			folders = append(folders, f.Folder)
		}
	}
	return folders, nil
}

package telegram

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/gotd/td/tg"
)

const tagPrefix = "#dir:"

func randID() int64 {
	var b [8]byte
	_, _ = rand.Read(b[:])
	return int64(binary.LittleEndian.Uint64(b[:]))
}

func extractMsgID(u tg.UpdatesClass) int {
	updates, ok := u.(*tg.Updates)
	if !ok {
		return 0
	}
	for _, upd := range updates.Updates {
		switch v := upd.(type) {
		case *tg.UpdateNewMessage:
			if m, ok := v.Message.(*tg.Message); ok {
				return m.ID
			}
		case *tg.UpdateMessageID:
			return v.ID
		}
	}
	return 0
}

func docFromMessage(ctx context.Context, api *tg.Client, msgID int) (*tg.Document, error) {
	res, err := api.MessagesGetMessages(ctx, []tg.InputMessageClass{
		&tg.InputMessageID{ID: msgID},
	})
	if err != nil {
		return nil, err
	}
	msgs, ok := res.(*tg.MessagesMessages)
	if !ok || len(msgs.Messages) == 0 {
		return nil, fmt.Errorf("message %d not found", msgID)
	}
	m, ok := msgs.Messages[0].(*tg.Message)
	if !ok {
		return nil, fmt.Errorf("message %d has no content", msgID)
	}
	media, ok := m.Media.(*tg.MessageMediaDocument)
	if !ok {
		return nil, fmt.Errorf("message %d has no document", msgID)
	}
	doc, ok := media.Document.(*tg.Document)
	if !ok {
		return nil, fmt.Errorf("message %d document unavailable", msgID)
	}
	return doc, nil
}

// tagCaption builds a caption like "#dir:Documents/photo.png" so the file
// is associated with the "Documents" folder. Root files get just the filename.
func tagCaption(folder, name string) string {
	if folder == "" || folder == "root" {
		return name
	}
	return tagPrefix + folder + "/" + name
}

// parseCaption extracts folder and display name from a message caption.
func parseCaption(caption string) (folder, name string) {
	if caption == "" {
		return "", ""
	}
	if strings.HasPrefix(caption, tagPrefix) {
		rest := strings.TrimPrefix(caption, tagPrefix)
		if idx := strings.LastIndex(rest, "/"); idx >= 0 {
			return rest[:idx], rest[idx+1:]
		}
		return rest, rest
	}
	return "", caption
}

func scanHistory(ctx context.Context, api *tg.Client, folder string, out *[]FileItem) error {
	wantRoot := folder == "" || folder == "root"
	wantAll := folder == "__all__"
	limit := 100
	offsetID := 0

	for {
		req := &tg.MessagesGetHistoryRequest{
			Peer:     &tg.InputPeerSelf{},
			OffsetID: offsetID,
			Limit:    limit,
		}
		res, err := api.MessagesGetHistory(ctx, req)
		if err != nil {
			return err
		}
		msgs, ok := res.(*tg.MessagesMessages)
		if !ok {
			break
		}
		if len(msgs.Messages) == 0 {
			break
		}
		for _, m := range msgs.Messages {
			msg, ok := m.(*tg.Message)
			if !ok {
				continue
			}
			fi := parseMessage(msg)
			if fi == nil {
				continue
			}
			offsetID = msg.ID

			if wantAll {
				*out = append(*out, *fi)
				continue
			}
			if wantRoot && fi.Folder == "" {
				*out = append(*out, *fi)
			} else if !wantRoot && fi.Folder == folder {
				*out = append(*out, *fi)
			}
		}
		if len(msgs.Messages) < limit {
			break
		}
	}
	return nil
}

func searchFiles(ctx context.Context, api *tg.Client, q, folder string, out *[]FileItem) error {
	res, err := api.MessagesSearch(ctx, &tg.MessagesSearchRequest{
		Peer:   &tg.InputPeerSelf{},
		Q:      q,
		Filter: &tg.InputMessagesFilterDocument{},
		Limit:  100,
	})
	if err != nil {
		return err
	}
	msgs, ok := res.(*tg.MessagesMessages)
	if !ok {
		return nil
	}
	wantRoot := folder == "" || folder == "root"
	for _, m := range msgs.Messages {
		msg, ok := m.(*tg.Message)
		if !ok {
			continue
		}
		fi := parseMessage(msg)
		if fi == nil {
			continue
		}
		if wantRoot && fi.Folder != "" {
			continue
		}
		if !wantRoot && fi.Folder != folder {
			continue
		}
		*out = append(*out, *fi)
	}
	return nil
}

func parseMessage(msg *tg.Message) *FileItem {
	media, ok := msg.Media.(*tg.MessageMediaDocument)
	if !ok {
		return nil
	}
	doc, ok := media.Document.(*tg.Document)
	if !ok {
		return nil
	}
	name := ""
	for _, attr := range doc.Attributes {
		if fn, ok := attr.(*tg.DocumentAttributeFilename); ok {
			name = fn.FileName
			break
		}
	}
	if name == "" {
		return nil
	}
	folder, displayName := parseCaption(msg.Message)
	return &FileItem{
		MsgID:    msg.ID,
		Name:     displayName,
		Size:     doc.Size,
		MimeType: doc.MimeType,
		Folder:   folder,
		Date:     int64(msg.Date),
	}
}

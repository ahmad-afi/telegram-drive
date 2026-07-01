package telegram

import (
	"bytes"
	"context"
	"fmt"
	"image/png"
	"sync"
	"time"

	"github.com/gotd/td/session"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth/qrlogin"
	"github.com/gotd/td/tg"
	qrcode "rsc.io/qr"
)

type QRManager struct {
	appID   int
	appHash string
	mu      sync.Mutex
	logins  map[string]*qrSession
}

type qrSession struct {
	qrImageCh chan string
	resultCh  chan qrResult
	cancel    context.CancelFunc
}

type qrResult struct {
	SessionB64 string
	Name       string
	Err        error
}

func NewQRManager(appID int, appHash string) *QRManager {
	return &QRManager{appID: appID, appHash: appHash, logins: make(map[string]*qrSession)}
}

func (m *QRManager) Start(loginID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)

	mem := &session.StorageMemory{}
	dispatcher := tg.NewUpdateDispatcher()
	client := telegram.NewClient(m.appID, m.appHash, telegram.Options{
		SessionStorage: mem,
		UpdateHandler:  dispatcher,
	})

	sess := &qrSession{
		qrImageCh: make(chan string, 1),
		resultCh:  make(chan qrResult, 1),
		cancel:    cancel,
	}
	m.mu.Lock()
	m.logins[loginID] = sess
	m.mu.Unlock()

	go func() {
		defer cancel()
		err := client.Run(ctx, func(ctx context.Context) error {
			loggedIn := qrlogin.OnLoginToken(&dispatcher)
			qr := client.QR()
			_, err := qr.Auth(ctx, loggedIn, func(ctx context.Context, token qrlogin.Token) error {
				img, err := token.Image(qrcode.M)
				if err != nil {
					return err
				}
				var buf bytes.Buffer
				if err := png.Encode(&buf, img); err != nil {
					return err
				}
				select {
				case sess.qrImageCh <- encodeBase64(buf.Bytes()):
				default:
				}
				return nil
			})
			if err != nil {
				return err
			}

			loader := &session.Loader{Storage: mem}
			d, err := loader.Load(ctx)
			if err != nil {
				return err
			}
			b64, err := EncodeSession(d)
			if err != nil {
				return err
			}
			self, err := client.Self(ctx)
			name := ""
			if err == nil && self.Username != "" {
				name = self.Username
			} else if err == nil {
				name = self.FirstName
			}
			sess.resultCh <- qrResult{SessionB64: b64, Name: name}
			return nil
		})
		if err != nil {
			sess.resultCh <- qrResult{Err: err}
		}
	}()

	return nil
}

// WaitForQR blocks briefly until the first QR image is ready, then returns it.
func (m *QRManager) WaitForQR(loginID string, timeout time.Duration) (string, error) {
	m.mu.Lock()
	sess, ok := m.logins[loginID]
	m.mu.Unlock()
	if !ok {
		return "", fmt.Errorf("login session not found")
	}
	select {
	case img := <-sess.qrImageCh:
		return img, nil
	case <-time.After(timeout):
		// Maybe already completed?
		select {
		case r := <-sess.resultCh:
			if r.Err != nil {
				return "", r.Err
			}
			return "", fmt.Errorf("login already completed")
		default:
		}
		return "", fmt.Errorf("QR image timeout")
	}
}

// Poll returns (session, name, true) if login succeeded, ("","",false) if pending.
func (m *QRManager) Poll(loginID string) (string, string, bool) {
	m.mu.Lock()
	sess, ok := m.logins[loginID]
	m.mu.Unlock()
	if !ok {
		return "", "", false
	}
	select {
	case r := <-sess.resultCh:
		if r.Err != nil {
			return "", "", false
		}
		return r.SessionB64, r.Name, true
	default:
		return "", "", false
	}
}

// Cancel stops a login session.
func (m *QRManager) Cancel(loginID string) {
	m.mu.Lock()
	sess, ok := m.logins[loginID]
	m.mu.Unlock()
	if ok {
		sess.cancel()
		m.mu.Lock()
		delete(m.logins, loginID)
		m.mu.Unlock()
	}
}

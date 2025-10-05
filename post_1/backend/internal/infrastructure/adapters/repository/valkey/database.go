package valkey_adapter

import (
	"personal_website/config"
	"personal_website/internal/app/core/ports"

	valkey "github.com/valkey-io/valkey-go"
)

type valkeyDatabase struct {
	client      valkey.Client
	sessionRepo ports.SessionRepository
}

func NewDatabase(cfg *config.ValkeyConfig) (*valkeyDatabase, error) {
	addr := cfg.Host.String() + ":" + cfg.Port.String()

	option := valkey.ClientOption{
		InitAddress: []string{addr},
	}

	if cfg.Password != nil && cfg.Password.String() != "" {
		option.Password = cfg.Password.String()
	}

	client, err := valkey.NewClient(option)
	if err != nil {
		return nil, err
	}

	sessionRepo := NewSessionAdapter(client)

	return &valkeyDatabase{
		client:      client,
		sessionRepo: sessionRepo,
	}, nil
}

func (d *valkeyDatabase) SessionRepo() ports.SessionRepository { return d.sessionRepo }

func (d *valkeyDatabase) Close() {
	d.client.Close()
}

package service

import (
	"fmt"

	"github.com/pgeowng/tamed/model"
	"github.com/pgeowng/tamed/store"
)

type MediaContentService interface {
	Upload() error
	Download() error
}

type MediaMetaService interface {
	Get(string) (*model.MediaMeta, error)
}

type Manager struct {
	MediaContent MediaContentService
	MediaMeta    MediaMetaService
}

func NewManager(store *store.Store) (*Manager, error) {
	if store == nil {
		return nil, fmt.Errorf("no store provided")
	} else {
		return &Manager{
			MediaContent: NewMediaContentSrv(store),
			MediaMeta:    NewMediaMetaSrv(store),
		}, nil
	}
}

package repo

import (
	"github.com/kabuke/ChroniclesFormosa/server/model"
)

// MapRepo 地圖庫介面 (未來可能是混合 Memory + SQLite Cache)
type MapRepo interface {
	SaveTile(tile *model.Tile) error
	GetChunk(startX, startY, width, height int32) ([]*model.Tile, error)
}

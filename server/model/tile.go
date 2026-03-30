package model

// Tile 地圖瓦片模型
type Tile struct {
	X       int32 `gorm:"primaryKey"`
	Y       int32 `gorm:"primaryKey"`
	Type    int32 // 0=草地, 1=山脈, 2=水域, 3=城鎮
	OwnerID int32 // 佔領者陣營ID
}

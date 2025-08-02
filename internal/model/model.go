package model

import (
	"github.com/google/uuid"
	"time"
)

type Banner struct {
	ID   uuid.UUID `db:"id"`
	Name string    `db:"name"`
}

// так как задача не большая, решил не писать отдельные структуры, для возврата из методов оставил тут json теги
type ClickStat struct {
	Timestamp time.Time `db:"timestamp" json:"timestamp"`
	BannerID  uuid.UUID `db:"banner_id" json:"banner_id"`
	Count     int       `db:"count" json:"count"`
}

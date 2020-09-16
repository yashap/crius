package dao

type Service struct {
	ID               int64  `gorm:"primaryKey"`
	Code             string `gorm:"index:idx_unique_code,unique;not null"`
	Name             string `gorm:"not null"`
	ServiceEndpoints []ServiceEndpoint
}

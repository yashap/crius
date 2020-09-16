package dao

type ServiceEndpointDependency struct {
	ID                          int64 `gorm:"primaryKey"`
	ServiceEndpointID           int64 `gorm:"index:idx_unique_endpoint_pair,unique,priority:1;not null"`
	DependencyServiceEndpointID int64 `gorm:"index:idx_unique_endpoint_pair,unique,priority:2;not null"`
}

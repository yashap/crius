package dao

type ServiceEndpoint struct {
	ID                          int64  `gorm:"primaryKey"`
	ServiceID                   int64  `gorm:"index:idx_unique_service_code_pair,unique,priority:1;not null"`
	Code                        string `gorm:"index:idx_unique_service_code_pair,unique,priority:2;not null"`
	Name                        string `gorm:"not null"`
	ServiceEndpointDependencies []ServiceEndpointDependency
}

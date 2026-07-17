package persistence

import (
	"context"

	"github.com/AbhijeetDev102/Nimbus/services/resource-service/internal/domain"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type postgresInstance struct {
	Client *gorm.DB
}

func NewPostgresInstance(dns string) (*postgresInstance, error) {

	db, err := gorm.Open(postgres.Open(dns), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	db.AutoMigrate(&domain.Resource{})

	return &postgresInstance{
		Client: db,
	}, nil
}

func (p *postgresInstance) CreateResource(ctx context.Context, resource *domain.Resource) error {

	err := p.Client.WithContext(ctx).Create(resource).Error
	if err != nil {
		return err
	}
	return nil
}

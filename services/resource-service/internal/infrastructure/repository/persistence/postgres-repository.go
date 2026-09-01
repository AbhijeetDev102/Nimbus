package persistence

import (
	"context"

	"github.com/AbhijeetDev102/Nimbus/services/resource-service/internal/domain"
	"github.com/google/uuid"
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

func (p *postgresInstance) GetResource(ctx context.Context, id uuid.UUID) (*domain.Resource, error) {
	var resource domain.Resource

	if err := p.Client.WithContext(ctx).First(&resource, "id=?", id).Error; err != nil {
		return nil, err
	}
	return &resource, nil
}

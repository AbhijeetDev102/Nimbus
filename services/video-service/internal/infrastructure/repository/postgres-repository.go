package repository

import (
	"context"

	"github.com/AbhijeetDev102/Nimbus/services/video-service/internal/domain"
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
	db.AutoMigrate(&domain.Video{})

	return &postgresInstance{
		Client: db,
	}, nil
}

func (p *postgresInstance) CreateVideo(ctx context.Context, video *domain.Video) error {

	err := p.Client.WithContext(ctx).Create(video).Error
	if err != nil {
		return err
	}
	return nil
}

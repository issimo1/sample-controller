package dao

import (
	"context"
	"fmt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Service struct {
	gorm.Model
	Name  string `gorm:"type:varchar(127);uniqueIndex"`
	Types string `gorm:"column:type"`
	Ip    string
}

func (s *Service) Create(ctx context.Context, service *Service) error {
	return DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Debug().Create(service).Error; err != nil {
			return fmt.Errorf("create service failed: %v", err)
		}
		return nil
	})
}

func (s *Service) Update(ctx context.Context, service *Service) error {
	return DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Model(service).Statement.Omit("type").Where("name = ? ", service.Name).Updates(service).Error; err != nil {
			return fmt.Errorf("update service failed: %v", err)
		}
		return nil
	})
}

func (s *Service) UpdateInBatchWithConflict(ctx context.Context, svc []*Service, batchSize int, column string, update []string) error {
	return DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Debug().Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: column},
			},
			DoUpdates: clause.AssignmentColumns(update),
		}).CreateInBatches(svc, batchSize).Error
		if err != nil {
			return fmt.Errorf("update service failed: %v", err)
		}
		return nil
	})
}

func (s *Service) Delete(ctx context.Context, service *Service) error {
	return DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(service).Error; err != nil {
			return fmt.Errorf("delete service failed: %v", err)
		}
		return nil
	})
}

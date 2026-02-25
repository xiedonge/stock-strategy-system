package services

import (
	"errors"

	"github.com/xiedonge/stock-strategy-system/backend/internal/models"
	"gorm.io/gorm"
)

// StrategyService manages strategy persistence.
type StrategyService struct {
	db *gorm.DB
}

// NewStrategyService constructs a StrategyService.
func NewStrategyService(db *gorm.DB) *StrategyService {
	return &StrategyService{db: db}
}

// List returns all strategies.
func (s *StrategyService) List() ([]models.Strategy, error) {
	var strategies []models.Strategy
	if err := s.db.Order("id desc").Find(&strategies).Error; err != nil {
		return nil, err
	}
	return strategies, nil
}

// Get fetches a strategy by ID.
func (s *StrategyService) Get(id uint) (*models.Strategy, error) {
	var strategy models.Strategy
	if err := s.db.First(&strategy, id).Error; err != nil {
		return nil, err
	}
	return &strategy, nil
}

// Create inserts a new strategy.
func (s *StrategyService) Create(strategy *models.Strategy) error {
	if strategy == nil {
		return errors.New("strategy is nil")
	}
	return s.db.Create(strategy).Error
}

// Update persists changes to a strategy.
func (s *StrategyService) Update(strategy *models.Strategy) error {
	if strategy == nil {
		return errors.New("strategy is nil")
	}
	return s.db.Save(strategy).Error
}

// Delete removes a strategy by ID.
func (s *StrategyService) Delete(id uint) error {
	return s.db.Delete(&models.Strategy{}, id).Error
}

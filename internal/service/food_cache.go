package service

import (
	"ai-calories/internal/model"
	"ai-calories/internal/repository"
)

type FoodCacheService struct {
	repo *repository.FoodCacheRepository
}

func NewFoodCacheService(repo *repository.FoodCacheRepository) *FoodCacheService {
	return &FoodCacheService{repo: repo}
}

func (s *FoodCacheService) Search(query string) ([]model.FoodCacheSearchResult, error) {
	items, err := s.repo.Search(query, 10)
	if err != nil {
		return nil, err
	}

	var results []model.FoodCacheSearchResult
	for _, item := range items {
		results = append(results, model.FoodCacheSearchResult{
			ID:           item.ID,
			FoodName:     item.FoodName,
			Calories100g: item.Calories100g,
			ImageURL:     item.ImageURL,
		})
	}
	return results, nil
}

func (s *FoodCacheService) GetByID(id uint) (*model.FoodCache, error) {
	return s.repo.FindByID(id)
}

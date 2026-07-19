package service

import (
	"ai-calories/internal/model"
	"ai-calories/internal/repository"
)

type UserService struct {
	userRepo *repository.UserRepository
}

func NewUserService(userRepo *repository.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

func (s *UserService) GetProfile(userID uint) (*model.UserProfileResponse, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}

	tz, _ := s.userRepo.GetTimezone(user.UserID)

	return &model.UserProfileResponse{
		ID:           user.ID,
		Email:        user.Email,
		AuthProvider: user.AuthProvider,
		Language:     user.Language,
		Timezone:     tz,
	}, nil
}

func (s *UserService) UpdateTimezone(userID int64, timezone string) error {
	return s.userRepo.UpdateTimezone(userID, timezone)
}

func (s *UserService) UpdateLanguage(id uint, language string) error {
	valid := map[string]bool{"en": true, "es-419": true, "pt-BR": true, "ru": true, "de": true, "fr": true}
	if !valid[language] {
		return &ValidationError{Field: "language", Message: "unsupported language"}
	}
	return s.userRepo.UpdateLanguage(id, language)
}

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Message
}

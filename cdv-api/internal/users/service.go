package users

import "context"

type Service struct {
	repository *Repository
}

func NewService(repository *Repository) *Service {
	return &Service{
		repository: repository,
	}
}

func (s *Service) GetUsers(ctx context.Context) ([]User, error) {
	return s.repository.GetAll(ctx)
}

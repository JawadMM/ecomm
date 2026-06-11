package account

import (
	"context"
	"github.com/segmentio/ksuid"
)

type Service struct {
	PostAccount  func(ctx context.Context, account *Account) (*Account, error)
	GetAccount   func(ctx context.Context, id string) (*Account, error)
	ListAccounts func(ctx context.Context, skip uint64, limit uint64) ([]Account, error)
}

type Account struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type AccountService struct {
	repository Repository
}

func NewService(repository Repository) *AccountService {
	return &AccountService{repository: repository}
}

func (s *AccountService) PostAccount(ctx context.Context, name string) (*Account, error) {
	account := &Account{
		ID:   ksuid.New().String(),
		Name: name,
	}
	if err := s.repository.PutAccount(ctx, account); err != nil {
		return nil, err
	}

	return account, nil
}

func (s *AccountService) GetAccount(ctx context.Context, id string) (*Account, error) {
	return s.repository.GetAccountByID(ctx, id)
}

func (s *AccountService) ListAccounts(ctx context.Context, skip uint64, take uint64) ([]Account, error) {
	if take > 100  || (skip ==0 && take == 0) {
		take = 100
	}
	return s.repository.ListAccounts(ctx, skip, take)
}

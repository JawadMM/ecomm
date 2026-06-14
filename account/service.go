package account

import (
	"context"
	"log"

	"github.com/JawadMM/ecomm/events"
	"github.com/segmentio/ksuid"
)

type Account struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type AccountService struct {
	repository Repository
	publisher  *events.Publisher
}

func NewService(repository Repository, pub *events.Publisher) *AccountService {
	return &AccountService{repository: repository, publisher: pub}
}

func (s *AccountService) PostAccount(ctx context.Context, name string) (*Account, error) {
	account := &Account{
		ID:   ksuid.New().String(),
		Name: name,
	}
	if err := s.repository.PutAccount(ctx, account); err != nil {
		return nil, err
	}
	if s.publisher != nil {
		s.publisher.Publish(events.SubjectAccountCreated, events.AccountCreatedEvent{
			ID:   account.ID,
			Name: account.Name,
		})
		log.Printf("[NATS pub] account.created — id=%s name=%s", account.ID, account.Name)
	}
	return account, nil
}

func (s *AccountService) GetAccount(ctx context.Context, id string) (*Account, error) {
	return s.repository.GetAccountByID(ctx, id)
}

func (s *AccountService) ListAccounts(ctx context.Context, skip uint64, take uint64) ([]Account, error) {
	if take > 100 || (skip == 0 && take == 0) {
		take = 100
	}
	return s.repository.ListAccounts(ctx, skip, take)
}

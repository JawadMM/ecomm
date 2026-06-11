package account

import (
	"context"
	"database/sql"

	_ "github.com/lib/pq"
)

type Repository interface {
	Close() error
	PutAccount(ctx context.Context, account *Account) error
	GetAccountByID(ctx context.Context, id string) (*Account, error)
	ListAccounts(ctx context.Context, skip uint64, take uint64) ([]Account, error)
}

type postgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(url string) (*postgresRepository, error) {
	db, err := sql.Open("postgres", url)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return &postgresRepository{db: db}, nil
}

func (r *postgresRepository) Close() error {
	return r.db.Close()
}

func (r *postgresRepository) Ping() error {
	return r.db.Ping()
}

func (r *postgresRepository) PutAccount(ctx context.Context, account *Account) error {
	_, err := r.db.ExecContext(ctx, "INSERT INTO accounts (id, name) VALUES ($1, $2)", account.ID, account.Name)
	return err
}

func (r *postgresRepository) GetAccountByID(ctx context.Context, id string) (*Account, error) {
	row := r.db.QueryRowContext(ctx, "SELECT id, name FROM accounts WHERE id = $1", id)
	account := Account{}

	err := row.Scan(&account.ID, &account.Name)

	if err != nil {
		return nil, err
	}
	return &account, nil
}

func (r *postgresRepository) ListAccounts(ctx context.Context, skip uint64, take uint64) ([]Account, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id, name FROM accounts ORDER BY id OFFSET $1 LIMIT $2", skip, take)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	accounts := []Account{}
	for rows.Next() {
		account := &Account{}
		err := rows.Scan(&account.ID, &account.Name)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, *account)
	}
	return accounts, nil
}

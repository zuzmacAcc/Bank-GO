package main

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

type Storage interface {
	CreateAccount(*Account) error
	DeleteAccount(int) error
	UpdateAccount(*Account) error
	GetAccounts() ([]*Account, error)
	GetAccountById(int) (*Account, error)
	GetAccountByNumber(int) (*Account, error)
	Deposit(int, int) error
	Transfer(int, int, int) error
}

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore() (*PostgresStore, error) {
	connStr := "user=postgres dbname=go-bank password=gobank port=5432 sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}
	return &PostgresStore{db: db}, nil
}

func (s *PostgresStore) Init() error {
	return s.createAccountTable()
}

func (s *PostgresStore) createAccountTable() error {
	query := `create table if not exists account (
		id serial primary key,
		first_name varchar(100),
		last_name varchar(100),
		number serial,
		encrypted_password varchar(100),
		balance serial,
		created_at timestamp
	)`

	_, err := s.db.Exec(query)
	return err
}

func (s *PostgresStore) CreateAccount(acc *Account) error {
	query := `insert into account
	(first_name, last_name, number, encrypted_password, balance, created_at)
	values ($1, $2, $3, $4, $5, $6)`

	_, err := s.db.Query(
		query,
		acc.FirstName,
		acc.LastName,
		acc.Number,
		acc.EncryptedPassword,
		acc.Balance,
		acc.CreatedAt)

	if err != nil {
		return err
	}

	return nil

	// rows, err := s.db.Query(
	// 	"select * from account where first_name = $1 and last_name = $2",
	// 	acc.FirstName,
	// 	acc.LastName)

	// if err != nil {
	// 	return nil, err
	// }

	// for rows.Next() {
	// 	account, err := scanIntoAccount(rows)

	// 	if err != nil {
	// 		return nil, err
	// 	}

	// 	return account, nil
	// }

	// return nil, fmt.Errorf("account %s %s not found", acc.FirstName, acc.LastName)
}

func (s *PostgresStore) UpdateAccount(*Account) error {
	return nil
}

func (s* PostgresStore) Transfer(amount, toAccount, id int) error {
	doubleUpdate(s.db, amount, toAccount, id)

	return nil
}

func (s* PostgresStore) Deposit(amount, id int) error {
	query := `update account set balance = balance + $1 where id = $2`

	_, err := s.db.Query(query, amount, id)

	if err != nil {
		return err
	}

	return nil
}

func (s *PostgresStore) DeleteAccount(id int) error {

	_, err := s.db.Query("delete from account where id = $1", id)

	return err
}

func (s *PostgresStore) GetAccountByNumber(number int) (*Account, error) {
	rows, err := s.db.Query("select * from account where number = $1", number)

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		return scanIntoAccount(rows)
	}

	return nil, fmt.Errorf("accoun with number <%d> not found", number)
}

func (s *PostgresStore) GetAccounts() ([]*Account, error) {
	rows, err := s.db.Query("select * from account")

	if err != nil {
		return nil, err
	}

	accounts := []*Account{}

	for rows.Next() {
		account, err := scanIntoAccount(rows)

		if err != nil {
			return nil, err
		}
		accounts = append(accounts, account)
	}

	return accounts, nil
}

func (s *PostgresStore) GetAccountById(id int) (*Account, error) {
	rows, err := s.db.Query("select * from account where id = $1", id)

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		account, err := scanIntoAccount(rows)

		if err != nil {
			return nil, err
		}

		return account, nil
	}

	return nil, fmt.Errorf("account %d not found", id)
}

func scanIntoAccount(rows *sql.Rows) (*Account, error) {
	account := new(Account)
	err := rows.Scan(
		&account.ID,
		&account.FirstName,
		&account.LastName,
		&account.Number,
		&account.EncryptedPassword,
		&account.Balance,
		&account.CreatedAt)

	return account, err
}

func doubleUpdate(db *sql.DB, amount, toAccount, id int) error {
	tx, err := db.Begin()
	if err != nil {
			return err
	}

	stmt, err := tx.Prepare(`update account set balance = balance - $1 where id = $2;`)
	if err != nil {
			tx.Rollback()
			return err
	}
	defer stmt.Close()

	if _, err := stmt.Exec(amount, id); err != nil {
			tx.Rollback() // return an error too, we may want to wrap them
			return err
	}

	stmt2, err := tx.Prepare(`update account set balance = balance + $1 where number = $2;`)
	if err != nil {
			tx.Rollback()
			return err
	}

	defer stmt2.Close()

	if _, err := stmt2.Exec(amount, toAccount); err != nil {
			tx.Rollback() // return an error too, we may want to wrap them
			return err
	}

	return tx.Commit()
}

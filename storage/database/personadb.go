package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

type PersonaDB struct {
	db *sql.DB
}

func InitDB(dbUser, dbName string) *PersonaDB {
	connStr := fmt.Sprintf("user=%s dbname=%s sslmode=disable", dbUser, dbName)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	return &PersonaDB{db}
}

func (p *PersonaDB) DB() *sql.DB {
	return p.db
}

func (p *PersonaDB) Close() {
	defer p.db.Close()
}

type Transaction interface {
	Begin(ctx context.Context) error
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

type MultiInstruction struct {
	db  *sql.DB
	tx  *sql.Tx
	ctx context.Context
}

func NewMultiInstruction(ctx context.Context, db *sql.DB) *MultiInstruction {
	return &MultiInstruction{
		db:  db,
		ctx: ctx,
		tx:  nil,
	}
}

func (t *MultiInstruction) Begin() error {
	tx, err := t.db.BeginTx(t.ctx, nil)
	if err != nil {
		return err
	}
	t.tx = tx
	return nil
}

func (t *MultiInstruction) Commit() error {
	return t.tx.Commit()
}

func (t *MultiInstruction) Rollback() error {
	return t.tx.Rollback()
}

func (t *MultiInstruction) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return t.tx.QueryContext(t.ctx, query, args...)
}

func (t *MultiInstruction) QueryRow(query string, args ...interface{}) *sql.Row {
	return t.tx.QueryRowContext(t.ctx, query, args...)
}

func (t *MultiInstruction) Exec(query string, args ...interface{}) (sql.Result, error) {
	return t.tx.ExecContext(t.ctx, query, args...)
}

func (t *MultiInstruction) Prepare(query string) (*sql.Stmt, error) {
	return t.tx.PrepareContext(t.ctx, query)
}

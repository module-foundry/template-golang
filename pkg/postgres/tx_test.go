package postgres

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
)

type fakeBeginner struct {
	tx  pgx.Tx
	err error
}

func (f fakeBeginner) Begin(context.Context) (pgx.Tx, error) { return f.tx, f.err }

type fakeTx struct {
	pgx.Tx
	commitErr   error
	committed   bool
	rolledBack  bool
	rollbackErr error
}

func (f *fakeTx) Commit(context.Context) error {
	f.committed = true
	return f.commitErr
}

func (f *fakeTx) Rollback(context.Context) error {
	f.rolledBack = true
	return f.rollbackErr
}

func TestWithTxSuccess(t *testing.T) {
	tx := &fakeTx{}
	called := false
	err := WithTx(context.Background(), fakeBeginner{tx: tx}, func(pgx.Tx) error {
		called = true
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called || !tx.committed {
		t.Fatal("fn must run and commit")
	}
}

func TestWithTxFnErrorRollsBack(t *testing.T) {
	tx := &fakeTx{}
	want := errors.New("boom")
	err := WithTx(context.Background(), fakeBeginner{tx: tx}, func(pgx.Tx) error {
		return want
	})
	if !errors.Is(err, want) {
		t.Fatalf("err = %v", err)
	}
	if !tx.rolledBack {
		t.Fatal("transaction must be rolled back")
	}
}

func TestWithTxBeginError(t *testing.T) {
	want := errors.New("no connection")
	err := WithTx(context.Background(), fakeBeginner{err: want}, func(pgx.Tx) error { return nil })
	if !errors.Is(err, want) {
		t.Fatalf("err = %v", err)
	}
}

func TestWithTxCommitError(t *testing.T) {
	want := errors.New("commit failed")
	tx := &fakeTx{commitErr: want}
	err := WithTx(context.Background(), fakeBeginner{tx: tx}, func(pgx.Tx) error { return nil })
	if !errors.Is(err, want) {
		t.Fatalf("err = %v", err)
	}
}

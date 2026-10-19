package postgres

import (
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/webitel/webitel-go-kit/infra/pgw"
	"github.com/webitel/webitel-go-kit/pkg/errors"
)

// NotNullColumn is a NOT NULL target registered with pgw.
type NotNullColumn struct {
	Table  string
	Column string
}

// Constraints names the Postgres constraints this service maps onto go-kit errors.
//
// pgw rewrites Exec and Query failures for 23502, 23503, 23505 and 23514 before
// the caller sees them. Registering each name makes that rewrite return ParseError,
// which still has the constraint name. An empty value means there is nothing to
// register yet.
type Constraints struct {
	Unique     []string
	ForeignKey []string
	Check      []string
	NotNull    []NotNullColumn
}

type constraintRegistry interface {
	RegisterUniqueViolation(constraintName string, processor pgw.ErrorProcessor) error
	RegisterForeignKeyViolation(constraintName string, processor pgw.ErrorProcessor) error
	RegisterCheckViolation(constraintName string, processor pgw.ErrorProcessor) error
	RegisterNotNullViolation(table, column string, processor pgw.ErrorProcessor) error
}

// RegisterConstraints installs go-kit processors for the named constraints.
func RegisterConstraints(manager *pgw.PoolManager, constraints Constraints) error {
	if manager == nil {
		return errors.New("postgres: pool manager is nil")
	}

	return applyConstraints(manager, constraints)
}

func applyConstraints(registry constraintRegistry, constraints Constraints) error {
	if err := registerNamed(constraints.Unique, registry.RegisterUniqueViolation); err != nil {
		return err
	}

	if err := registerNamed(constraints.ForeignKey, registry.RegisterForeignKeyViolation); err != nil {
		return err
	}

	if err := registerNamed(constraints.Check, registry.RegisterCheckViolation); err != nil {
		return err
	}

	return registerNotNull(constraints.NotNull, registry)
}

func registerNamed(names []string, register func(string, pgw.ErrorProcessor) error) error {
	for _, name := range names {
		if err := registerOne(name, register); err != nil {
			return err
		}
	}

	return nil
}

func registerOne(name string, register func(string, pgw.ErrorProcessor) error) error {
	if name == "" {
		return errors.New("postgres: constraint name is empty")
	}

	if err := register(name, pgErrorProcessor); err != nil {
		return fmt.Errorf("postgres: register constraint %s: %w", name, err)
	}

	return nil
}

func registerNotNull(columns []NotNullColumn, registry constraintRegistry) error {
	for _, column := range columns {
		if err := registerOneNotNull(column, registry); err != nil {
			return err
		}
	}

	return nil
}

func registerOneNotNull(column NotNullColumn, registry constraintRegistry) error {
	if column.Table == "" || column.Column == "" {
		return errors.New("postgres: not-null table and column are required")
	}

	err := registry.RegisterNotNullViolation(column.Table, column.Column, pgErrorProcessor)
	if err != nil {
		return fmt.Errorf("postgres: register not-null %s.%s: %w", column.Table, column.Column, err)
	}

	return nil
}

func pgErrorProcessor(pgErr *pgconn.PgError) error {
	if pgErr == nil {
		return nil
	}

	return ParseError(pgErr)
}

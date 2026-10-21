package postgres

import (
	"errors"
	"strings"
	"testing"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"google.golang.org/grpc/codes"

	"github.com/webitel/webitel-go-kit/infra/pgw"
	kiterrors "github.com/webitel/webitel-go-kit/pkg/errors"
)

func TestRegisterConstraintsRejectsNilManager(t *testing.T) {
	err := RegisterConstraints(nil, Constraints{})
	if err == nil || !strings.Contains(err.Error(), "pool manager is nil") {
		t.Fatalf("RegisterConstraints() error = %v", err)
	}
}

func TestApplyConstraints(t *testing.T) {
	registry := &fakeRegistry{}
	constraints := Constraints{
		Unique:     []string{"space_name_uq"},
		ForeignKey: []string{"space_model_fk"},
		Check:      []string{"dim_check"},
		NotNull:    []NotNullColumn{{Table: "space", Column: "name"}},
	}

	if err := applyConstraints(registry, constraints); err != nil {
		t.Fatalf("applyConstraints() error = %v", err)
	}

	if got, want := strings.Join(registry.unique, ","), "space_name_uq"; got != want {
		t.Fatalf("unique = %s", got)
	}

	if got, want := strings.Join(registry.foreignKey, ","), "space_model_fk"; got != want {
		t.Fatalf("foreign key = %s", got)
	}

	if got, want := strings.Join(registry.check, ","), "dim_check"; got != want {
		t.Fatalf("check = %s", got)
	}

	if got, want := strings.Join(registry.notNull, ","), "space.name"; got != want {
		t.Fatalf("not null = %s", got)
	}

	got := registry.processor(&pgconn.PgError{
		Code:           pgerrcode.UniqueViolation,
		ConstraintName: "space_name_uq",
	})
	if kiterrors.Code(got) != codes.AlreadyExists {
		t.Fatalf("processor code = %v", kiterrors.Code(got))
	}

	if value, _ := kiterrors.Value(got, "constraint").(string); value != "space_name_uq" {
		t.Fatalf("processor constraint = %q", value)
	}
}

func TestApplyConstraintsRejectsEmptyName(t *testing.T) {
	registry := &fakeRegistry{}

	err := applyConstraints(registry, Constraints{Unique: []string{""}})
	if err == nil || !strings.Contains(err.Error(), "constraint name is empty") {
		t.Fatalf("applyConstraints() error = %v", err)
	}

	if len(registry.unique) != 0 {
		t.Fatalf("registered %v", registry.unique)
	}
}

func TestApplyConstraintsRejectsEmptyNotNull(t *testing.T) {
	err := applyConstraints(&fakeRegistry{}, Constraints{NotNull: []NotNullColumn{{Table: "space"}}})
	if err == nil || !strings.Contains(err.Error(), "not-null table and column are required") {
		t.Fatalf("applyConstraints() error = %v", err)
	}
}

func TestApplyConstraintsStopsOnRegisterError(t *testing.T) {
	registry := &fakeRegistry{failOn: "space_name_uq"}

	err := applyConstraints(registry, Constraints{
		Unique:     []string{"space_name_uq"},
		ForeignKey: []string{"space_model_fk"},
	})
	if err == nil || !strings.Contains(err.Error(), "register constraint space_name_uq") {
		t.Fatalf("applyConstraints() error = %v", err)
	}

	if len(registry.foreignKey) != 0 {
		t.Fatalf("foreign key registered after failure: %v", registry.foreignKey)
	}
}

func TestPgErrorProcessorNil(t *testing.T) {
	if err := pgErrorProcessor(nil); err != nil {
		t.Fatalf("pgErrorProcessor(nil) = %v", err)
	}
}

type fakeRegistry struct {
	unique     []string
	foreignKey []string
	check      []string
	notNull    []string
	processor  pgw.ErrorProcessor
	failOn     string
}

func (f *fakeRegistry) RegisterUniqueViolation(name string, processor pgw.ErrorProcessor) error {
	return f.record(&f.unique, name, processor)
}

func (f *fakeRegistry) RegisterForeignKeyViolation(name string, processor pgw.ErrorProcessor) error {
	return f.record(&f.foreignKey, name, processor)
}

func (f *fakeRegistry) RegisterCheckViolation(name string, processor pgw.ErrorProcessor) error {
	return f.record(&f.check, name, processor)
}

func (f *fakeRegistry) RegisterNotNullViolation(table, column string, processor pgw.ErrorProcessor) error {
	return f.record(&f.notNull, table+"."+column, processor)
}

func (f *fakeRegistry) record(dst *[]string, name string, processor pgw.ErrorProcessor) error {
	if name == f.failOn {
		return errors.New("boom")
	}

	*dst = append(*dst, name)
	f.processor = processor

	return nil
}

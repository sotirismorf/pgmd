package markdown

import (
	"strings"
	"testing"

	"github.com/sotirismorf/pgmd/internal/pg"
)

func TestRender_EmptySchema(t *testing.T) {
	schemas := []pg.SchemaInfo{
		{Name: "public"},
	}

	result := Render(schemas)

	if !strings.Contains(result, "# Database Schema Documentation") {
		t.Error("expected header not found")
	}
	if !strings.Contains(result, "## Schema: public") {
		t.Error("expected schema name not found")
	}
}

func TestRender_TableWithColumns(t *testing.T) {
	schemas := []pg.SchemaInfo{
		{
			Name: "public",
			Tables: []pg.Table{
				{
					Schema: "public",
					Name:   "users",
					Columns: []pg.Column{
						{Name: "id", Type: "uuid", Nullable: false, IsPK: true},
						{Name: "email", Type: "text", Nullable: false, IsUnique: true},
						{Name: "name", Type: "text", Nullable: true},
					},
				},
			},
		},
	}

	result := Render(schemas)

	if !strings.Contains(result, "### Tables") {
		t.Error("expected Tables section not found")
	}
	if !strings.Contains(result, "#### users") {
		t.Error("expected table name not found")
	}
	if !strings.Contains(result, "| id | uuid | PK, NOT NULL |") {
		t.Error("expected PK column not found")
	}
	if !strings.Contains(result, "| email | text | NOT NULL, UNIQUE |") {
		t.Error("expected unique column not found")
	}
	if !strings.Contains(result, "| name | text |  |") {
		t.Error("expected nullable column not found")
	}
}

func TestRender_TableWithForeignKey(t *testing.T) {
	schemas := []pg.SchemaInfo{
		{
			Name: "public",
			Tables: []pg.Table{
				{
					Schema: "public",
					Name:   "posts",
					Columns: []pg.Column{
						{Name: "user_id", Type: "uuid", Nullable: false, FKRef: "public.users.id"},
					},
				},
			},
		},
	}

	result := Render(schemas)

	if !strings.Contains(result, "FK→public.users.id") {
		t.Error("expected FK reference not found")
	}
}

func TestRender_TableWithIndexes(t *testing.T) {
	schemas := []pg.SchemaInfo{
		{
			Name: "public",
			Tables: []pg.Table{
				{
					Schema: "public",
					Name:   "users",
					Columns: []pg.Column{
						{Name: "id", Type: "uuid", IsPK: true},
					},
					Indexes: []pg.Index{
						{Name: "users_pkey", Columns: []string{"id"}, IsPrimary: true},
						{Name: "idx_email", Columns: []string{"email"}, IsUnique: true},
					},
				},
			},
		},
	}

	result := Render(schemas)

	if !strings.Contains(result, "**Indexes:**") {
		t.Error("expected Indexes section not found")
	}
	if !strings.Contains(result, "users_pkey (id, PK)") {
		t.Error("expected primary key index not found")
	}
	if !strings.Contains(result, "idx_email (email, UNIQUE)") {
		t.Error("expected unique index not found")
	}
}

func TestRender_Views(t *testing.T) {
	schemas := []pg.SchemaInfo{
		{
			Name: "public",
			Views: []pg.View{
				{
					Schema: "public",
					Name:   "active_users",
					Columns: []pg.Column{
						{Name: "id", Type: "uuid"},
						{Name: "email", Type: "text"},
					},
				},
			},
		},
	}

	result := Render(schemas)

	if !strings.Contains(result, "### Views") {
		t.Error("expected Views section not found")
	}
	if !strings.Contains(result, "#### active_users") {
		t.Error("expected view name not found")
	}
}

func TestRender_MaterializedViews(t *testing.T) {
	schemas := []pg.SchemaInfo{
		{
			Name: "public",
			MaterializedViews: []pg.MaterializedView{
				{
					Schema: "public",
					Name:   "user_stats",
					Columns: []pg.Column{
						{Name: "user_id", Type: "uuid"},
						{Name: "post_count", Type: "bigint"},
					},
				},
			},
		},
	}

	result := Render(schemas)

	if !strings.Contains(result, "### Materialized Views") {
		t.Error("expected Materialized Views section not found")
	}
	if !strings.Contains(result, "#### user_stats") {
		t.Error("expected materialized view name not found")
	}
}

func TestRender_Sequences(t *testing.T) {
	schemas := []pg.SchemaInfo{
		{
			Name: "public",
			Sequences: []pg.Sequence{
				{
					Schema:    "public",
					Name:      "users_id_seq",
					DataType:  "bigint",
					Start:     1,
					Min:       1,
					Max:       9223372036854775807,
					Increment: 1,
					Cycle:     false,
				},
			},
		},
	}

	result := Render(schemas)

	if !strings.Contains(result, "### Sequences") {
		t.Error("expected Sequences section not found")
	}
	if !strings.Contains(result, "`users_id_seq` (bigint)") {
		t.Error("expected sequence name not found")
	}
	if !strings.Contains(result, "start=1") {
		t.Error("expected start value not found")
	}
}

func TestRender_SequenceWithCycle(t *testing.T) {
	schemas := []pg.SchemaInfo{
		{
			Name: "public",
			Sequences: []pg.Sequence{
				{
					Schema:    "public",
					Name:      "rotating_seq",
					DataType:  "integer",
					Start:     1,
					Min:       1,
					Max:       100,
					Increment: 1,
					Cycle:     true,
				},
			},
		},
	}

	result := Render(schemas)

	if !strings.Contains(result, ", CYCLE") {
		t.Error("expected CYCLE flag not found")
	}
}

func TestRender_Triggers(t *testing.T) {
	schemas := []pg.SchemaInfo{
		{
			Name: "public",
			Triggers: []pg.Trigger{
				{
					Schema:   "public",
					Table:    "users",
					Name:     "update_timestamp",
					Event:    "UPDATE",
					Timing:   "BEFORE",
					Function: "set_updated_at",
				},
			},
		},
	}

	result := Render(schemas)

	if !strings.Contains(result, "### Triggers") {
		t.Error("expected Triggers section not found")
	}
	if !strings.Contains(result, "`update_timestamp` on `users`") {
		t.Error("expected trigger name not found")
	}
	if !strings.Contains(result, "BEFORE UPDATE") {
		t.Error("expected trigger timing/event not found")
	}
	if !strings.Contains(result, "set_updated_at()") {
		t.Error("expected trigger function not found")
	}
}

func TestRender_Functions(t *testing.T) {
	schemas := []pg.SchemaInfo{
		{
			Name: "public",
			Functions: []pg.Function{
				{
					Schema:     "public",
					Name:       "get_user",
					Arguments:  "id uuid",
					ReturnType: "users",
				},
				{
					Schema:     "public",
					Name:       "count_users",
					Arguments:  "",
					ReturnType: "bigint",
				},
			},
		},
	}

	result := Render(schemas)

	if !strings.Contains(result, "### Functions") {
		t.Error("expected Functions section not found")
	}
	if !strings.Contains(result, "`get_user(id uuid) → users`") {
		t.Error("expected function with args not found")
	}
	if !strings.Contains(result, "`count_users() → bigint`") {
		t.Error("expected function without args not found")
	}
}

func TestRender_CustomTypes(t *testing.T) {
	schemas := []pg.SchemaInfo{
		{
			Name: "public",
			Types: []pg.CustomType{
				{
					Schema: "public",
					Name:   "status",
					Kind:   "enum",
					Values: []string{"pending", "active", "archived"},
				},
				{
					Schema: "public",
					Name:   "address",
					Kind:   "composite",
					Values: []string{"street text", "city text", "zip text"},
				},
			},
		},
	}

	result := Render(schemas)

	if !strings.Contains(result, "### Custom Types") {
		t.Error("expected Custom Types section not found")
	}
	if !strings.Contains(result, "`status`: 'pending', 'active', 'archived'") {
		t.Error("expected enum type not found")
	}
	if !strings.Contains(result, "`address` (composite)") {
		t.Error("expected composite type not found")
	}
}

func TestRender_MultipleSchemas(t *testing.T) {
	schemas := []pg.SchemaInfo{
		{Name: "public"},
		{Name: "auth"},
	}

	result := Render(schemas)

	if !strings.Contains(result, "## Schema: public") {
		t.Error("expected public schema not found")
	}
	if !strings.Contains(result, "## Schema: auth") {
		t.Error("expected auth schema not found")
	}
	if !strings.Contains(result, "---") {
		t.Error("expected schema separator not found")
	}
}

func TestBuildConstraints(t *testing.T) {
	tests := []struct {
		name     string
		col      pg.Column
		expected string
	}{
		{
			name:     "primary key",
			col:      pg.Column{IsPK: true, Nullable: false},
			expected: "PK, NOT NULL",
		},
		{
			name:     "unique not null",
			col:      pg.Column{IsUnique: true, Nullable: false},
			expected: "NOT NULL, UNIQUE",
		},
		{
			name:     "foreign key",
			col:      pg.Column{FKRef: "public.users.id", Nullable: false},
			expected: "NOT NULL, FK→public.users.id",
		},
		{
			name:     "with default",
			col:      pg.Column{Default: "now()", Nullable: false},
			expected: "NOT NULL, DEFAULT now()",
		},
		{
			name:     "nullable no constraints",
			col:      pg.Column{Nullable: true},
			expected: "",
		},
		{
			name:     "pk is also unique - no duplicate",
			col:      pg.Column{IsPK: true, IsUnique: true, Nullable: false},
			expected: "PK, NOT NULL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildConstraints(tt.col)
			if result != tt.expected {
				t.Errorf("buildConstraints() = %q, want %q", result, tt.expected)
			}
		})
	}
}

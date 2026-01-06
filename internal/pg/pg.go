package pg

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
)

type Column struct {
	Name     string
	Type     string
	Nullable bool
	IsPK     bool
	IsUnique bool
	FKRef    string
	Default  string
}

type Index struct {
	Name      string
	Columns   []string
	IsUnique  bool
	IsPrimary bool
}

type Table struct {
	Schema  string
	Name    string
	Columns []Column
	Indexes []Index
}

type View struct {
	Schema  string
	Name    string
	Columns []Column
}

type Function struct {
	Schema     string
	Name       string
	Arguments  string
	ReturnType string
}

type CustomType struct {
	Schema string
	Name   string
	Kind   string
	Values []string
}

type MaterializedView struct {
	Schema  string
	Name    string
	Columns []Column
}

type Sequence struct {
	Schema    string
	Name      string
	DataType  string
	Start     int64
	Min       int64
	Max       int64
	Increment int64
	Cycle     bool
}

type Trigger struct {
	Schema    string
	Table     string
	Name      string
	Event     string
	Timing    string
	Function  string
}

type SchemaInfo struct {
	Name              string
	Tables            []Table
	Views             []View
	MaterializedViews []MaterializedView
	Sequences         []Sequence
	Triggers          []Trigger
	Functions         []Function
	Types             []CustomType
}

func FetchSchemas(ctx context.Context, conn *pgx.Conn, schemas []string) ([]SchemaInfo, error) {
	var result []SchemaInfo

	for _, schema := range schemas {
		info := SchemaInfo{Name: schema}

		tables, err := fetchTables(ctx, conn, schema)
		if err != nil {
			return nil, fmt.Errorf("fetching tables for schema %s: %w", schema, err)
		}
		info.Tables = tables

		views, err := fetchViews(ctx, conn, schema)
		if err != nil {
			return nil, fmt.Errorf("fetching views for schema %s: %w", schema, err)
		}
		info.Views = views

		matViews, err := fetchMaterializedViews(ctx, conn, schema)
		if err != nil {
			return nil, fmt.Errorf("fetching materialized views for schema %s: %w", schema, err)
		}
		info.MaterializedViews = matViews

		sequences, err := fetchSequences(ctx, conn, schema)
		if err != nil {
			return nil, fmt.Errorf("fetching sequences for schema %s: %w", schema, err)
		}
		info.Sequences = sequences

		triggers, err := fetchTriggers(ctx, conn, schema)
		if err != nil {
			return nil, fmt.Errorf("fetching triggers for schema %s: %w", schema, err)
		}
		info.Triggers = triggers

		functions, err := fetchFunctions(ctx, conn, schema)
		if err != nil {
			return nil, fmt.Errorf("fetching functions for schema %s: %w", schema, err)
		}
		info.Functions = functions

		types, err := fetchCustomTypes(ctx, conn, schema)
		if err != nil {
			return nil, fmt.Errorf("fetching types for schema %s: %w", schema, err)
		}
		info.Types = types

		result = append(result, info)
	}

	return result, nil
}

func fetchTables(ctx context.Context, conn *pgx.Conn, schema string) ([]Table, error) {
	query := `
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema = $1
		  AND table_type = 'BASE TABLE'
		ORDER BY table_name`

	rows, err := conn.Query(ctx, query, schema)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []Table
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		tables = append(tables, Table{Schema: schema, Name: name})
	}

	for i := range tables {
		columns, err := fetchColumns(ctx, conn, schema, tables[i].Name)
		if err != nil {
			return nil, err
		}
		tables[i].Columns = columns

		indexes, err := fetchIndexes(ctx, conn, schema, tables[i].Name)
		if err != nil {
			return nil, err
		}
		tables[i].Indexes = indexes
	}

	return tables, nil
}

func fetchColumns(ctx context.Context, conn *pgx.Conn, schema, table string) ([]Column, error) {
	query := `
		SELECT
			c.column_name,
			c.data_type,
			c.is_nullable,
			c.column_default,
			COALESCE(
				(SELECT true FROM information_schema.table_constraints tc
				 JOIN information_schema.key_column_usage kcu
				   ON tc.constraint_name = kcu.constraint_name
				  AND tc.table_schema = kcu.table_schema
				 WHERE tc.constraint_type = 'PRIMARY KEY'
				   AND tc.table_schema = c.table_schema
				   AND tc.table_name = c.table_name
				   AND kcu.column_name = c.column_name
				 LIMIT 1), false) as is_pk,
			COALESCE(
				(SELECT true FROM information_schema.table_constraints tc
				 JOIN information_schema.key_column_usage kcu
				   ON tc.constraint_name = kcu.constraint_name
				  AND tc.table_schema = kcu.table_schema
				 WHERE tc.constraint_type = 'UNIQUE'
				   AND tc.table_schema = c.table_schema
				   AND tc.table_name = c.table_name
				   AND kcu.column_name = c.column_name
				 LIMIT 1), false) as is_unique,
			COALESCE(
				(SELECT ccu.table_schema || '.' || ccu.table_name || '.' || ccu.column_name
				 FROM information_schema.table_constraints tc
				 JOIN information_schema.key_column_usage kcu
				   ON tc.constraint_name = kcu.constraint_name
				  AND tc.table_schema = kcu.table_schema
				 JOIN information_schema.constraint_column_usage ccu
				   ON tc.constraint_name = ccu.constraint_name
				  AND tc.table_schema = ccu.table_schema
				 WHERE tc.constraint_type = 'FOREIGN KEY'
				   AND tc.table_schema = c.table_schema
				   AND tc.table_name = c.table_name
				   AND kcu.column_name = c.column_name
				 LIMIT 1), '') as fk_ref
		FROM information_schema.columns c
		WHERE c.table_schema = $1
		  AND c.table_name = $2
		ORDER BY c.ordinal_position`

	rows, err := conn.Query(ctx, query, schema, table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []Column
	for rows.Next() {
		var col Column
		var nullable string
		var defaultVal *string

		if err := rows.Scan(&col.Name, &col.Type, &nullable, &defaultVal, &col.IsPK, &col.IsUnique, &col.FKRef); err != nil {
			return nil, err
		}

		col.Nullable = nullable == "YES"
		if defaultVal != nil {
			col.Default = *defaultVal
		}

		columns = append(columns, col)
	}

	return columns, nil
}

func fetchIndexes(ctx context.Context, conn *pgx.Conn, schema, table string) ([]Index, error) {
	query := `
		SELECT
			i.relname as index_name,
			array_agg(a.attname ORDER BY array_position(ix.indkey, a.attnum)) as columns,
			ix.indisunique as is_unique,
			ix.indisprimary as is_primary
		FROM pg_index ix
		JOIN pg_class i ON i.oid = ix.indexrelid
		JOIN pg_class t ON t.oid = ix.indrelid
		JOIN pg_namespace n ON n.oid = t.relnamespace
		JOIN pg_attribute a ON a.attrelid = t.oid AND a.attnum = ANY(ix.indkey)
		WHERE n.nspname = $1
		  AND t.relname = $2
		GROUP BY i.relname, ix.indisunique, ix.indisprimary
		ORDER BY i.relname`

	rows, err := conn.Query(ctx, query, schema, table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var indexes []Index
	for rows.Next() {
		var idx Index
		if err := rows.Scan(&idx.Name, &idx.Columns, &idx.IsUnique, &idx.IsPrimary); err != nil {
			return nil, err
		}
		indexes = append(indexes, idx)
	}

	return indexes, nil
}

func fetchViews(ctx context.Context, conn *pgx.Conn, schema string) ([]View, error) {
	query := `
		SELECT table_name
		FROM information_schema.views
		WHERE table_schema = $1
		ORDER BY table_name`

	rows, err := conn.Query(ctx, query, schema)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var views []View
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		views = append(views, View{Schema: schema, Name: name})
	}

	for i := range views {
		columns, err := fetchViewColumns(ctx, conn, schema, views[i].Name)
		if err != nil {
			return nil, err
		}
		views[i].Columns = columns
	}

	return views, nil
}

func fetchViewColumns(ctx context.Context, conn *pgx.Conn, schema, view string) ([]Column, error) {
	query := `
		SELECT
			column_name,
			data_type,
			is_nullable
		FROM information_schema.columns
		WHERE table_schema = $1
		  AND table_name = $2
		ORDER BY ordinal_position`

	rows, err := conn.Query(ctx, query, schema, view)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []Column
	for rows.Next() {
		var col Column
		var nullable string

		if err := rows.Scan(&col.Name, &col.Type, &nullable); err != nil {
			return nil, err
		}

		col.Nullable = nullable == "YES"
		columns = append(columns, col)
	}

	return columns, nil
}

func fetchFunctions(ctx context.Context, conn *pgx.Conn, schema string) ([]Function, error) {
	query := `
		SELECT
			p.proname as name,
			pg_get_function_arguments(p.oid) as arguments,
			pg_get_function_result(p.oid) as return_type
		FROM pg_proc p
		JOIN pg_namespace n ON n.oid = p.pronamespace
		WHERE n.nspname = $1
		  AND p.prokind = 'f'
		ORDER BY p.proname`

	rows, err := conn.Query(ctx, query, schema)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var functions []Function
	for rows.Next() {
		var fn Function
		fn.Schema = schema
		if err := rows.Scan(&fn.Name, &fn.Arguments, &fn.ReturnType); err != nil {
			return nil, err
		}
		functions = append(functions, fn)
	}

	return functions, nil
}

func fetchCustomTypes(ctx context.Context, conn *pgx.Conn, schema string) ([]CustomType, error) {
	var types []CustomType

	// Fetch enums
	enumQuery := `
		SELECT t.typname, array_agg(e.enumlabel ORDER BY e.enumsortorder)
		FROM pg_type t
		JOIN pg_namespace n ON n.oid = t.typnamespace
		JOIN pg_enum e ON e.enumtypid = t.oid
		WHERE n.nspname = $1
		GROUP BY t.typname
		ORDER BY t.typname`

	rows, err := conn.Query(ctx, enumQuery, schema)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var ct CustomType
		ct.Schema = schema
		ct.Kind = "enum"
		if err := rows.Scan(&ct.Name, &ct.Values); err != nil {
			return nil, err
		}
		types = append(types, ct)
	}

	// Fetch composite types
	compositeQuery := `
		SELECT t.typname,
			   array_agg(a.attname || ' ' || pg_catalog.format_type(a.atttypid, a.atttypmod) ORDER BY a.attnum)
		FROM pg_type t
		JOIN pg_namespace n ON n.oid = t.typnamespace
		JOIN pg_class c ON c.oid = t.typrelid
		JOIN pg_attribute a ON a.attrelid = c.oid AND a.attnum > 0 AND NOT a.attisdropped
		WHERE n.nspname = $1
		  AND t.typtype = 'c'
		  AND c.relkind = 'c'
		GROUP BY t.typname
		ORDER BY t.typname`

	rows2, err := conn.Query(ctx, compositeQuery, schema)
	if err != nil {
		return nil, err
	}
	defer rows2.Close()

	for rows2.Next() {
		var ct CustomType
		ct.Schema = schema
		ct.Kind = "composite"
		if err := rows2.Scan(&ct.Name, &ct.Values); err != nil {
			return nil, err
		}
		types = append(types, ct)
	}

	return types, nil
}

func ParseSchemas(input string) []string {
	parts := strings.Split(input, ",")
	var schemas []string
	for _, p := range parts {
		s := strings.TrimSpace(p)
		if s != "" {
			schemas = append(schemas, s)
		}
	}
	return schemas
}

func fetchMaterializedViews(ctx context.Context, conn *pgx.Conn, schema string) ([]MaterializedView, error) {
	query := `
		SELECT matviewname
		FROM pg_matviews
		WHERE schemaname = $1
		ORDER BY matviewname`

	rows, err := conn.Query(ctx, query, schema)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var views []MaterializedView
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		views = append(views, MaterializedView{Schema: schema, Name: name})
	}

	for i := range views {
		columns, err := fetchViewColumns(ctx, conn, schema, views[i].Name)
		if err != nil {
			return nil, err
		}
		views[i].Columns = columns
	}

	return views, nil
}

func fetchSequences(ctx context.Context, conn *pgx.Conn, schema string) ([]Sequence, error) {
	query := `
		SELECT
			sequencename,
			data_type::text,
			start_value,
			min_value,
			max_value,
			increment_by,
			cycle
		FROM pg_sequences
		WHERE schemaname = $1
		ORDER BY sequencename`

	rows, err := conn.Query(ctx, query, schema)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sequences []Sequence
	for rows.Next() {
		var seq Sequence
		seq.Schema = schema
		if err := rows.Scan(&seq.Name, &seq.DataType, &seq.Start, &seq.Min, &seq.Max, &seq.Increment, &seq.Cycle); err != nil {
			return nil, err
		}
		sequences = append(sequences, seq)
	}

	return sequences, nil
}

func fetchTriggers(ctx context.Context, conn *pgx.Conn, schema string) ([]Trigger, error) {
	query := `
		SELECT
			c.relname as table_name,
			t.tgname as trigger_name,
			CASE
				WHEN t.tgtype & 2 = 2 THEN 'BEFORE'
				WHEN t.tgtype & 2 = 0 THEN 'AFTER'
				ELSE 'INSTEAD OF'
			END as timing,
			array_to_string(ARRAY[
				CASE WHEN t.tgtype & 4 = 4 THEN 'INSERT' END,
				CASE WHEN t.tgtype & 8 = 8 THEN 'DELETE' END,
				CASE WHEN t.tgtype & 16 = 16 THEN 'UPDATE' END,
				CASE WHEN t.tgtype & 32 = 32 THEN 'TRUNCATE' END
			]::text[], ' OR ') as event,
			p.proname as function_name
		FROM pg_trigger t
		JOIN pg_class c ON c.oid = t.tgrelid
		JOIN pg_namespace n ON n.oid = c.relnamespace
		JOIN pg_proc p ON p.oid = t.tgfoid
		WHERE n.nspname = $1
		  AND NOT t.tgisinternal
		ORDER BY c.relname, t.tgname`

	rows, err := conn.Query(ctx, query, schema)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var triggers []Trigger
	for rows.Next() {
		var trig Trigger
		trig.Schema = schema
		if err := rows.Scan(&trig.Table, &trig.Name, &trig.Timing, &trig.Event, &trig.Function); err != nil {
			return nil, err
		}
		triggers = append(triggers, trig)
	}

	return triggers, nil
}

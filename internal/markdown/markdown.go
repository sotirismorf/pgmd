package markdown

import (
	"fmt"
	"strings"

	"github.com/sotirismorf/pgmd/internal/pg"
)

func Render(schemas []pg.SchemaInfo) string {
	var sb strings.Builder

	sb.WriteString("# Database Schema Documentation\n\n")

	for i, schema := range schemas {
		if i > 0 {
			sb.WriteString("\n---\n\n")
		}
		renderSchema(&sb, schema)
	}

	return sb.String()
}

func renderSchema(sb *strings.Builder, schema pg.SchemaInfo) {
	fmt.Fprintf(sb, "## Schema: %s\n\n", schema.Name)

	if len(schema.Tables) > 0 {
		sb.WriteString("### Tables\n\n")
		for _, table := range schema.Tables {
			renderTable(sb, table)
		}
	}

	if len(schema.Views) > 0 {
		sb.WriteString("### Views\n\n")
		for _, view := range schema.Views {
			renderView(sb, view)
		}
	}

	if len(schema.MaterializedViews) > 0 {
		sb.WriteString("### Materialized Views\n\n")
		for _, mv := range schema.MaterializedViews {
			renderMaterializedView(sb, mv)
		}
	}

	if len(schema.Sequences) > 0 {
		sb.WriteString("### Sequences\n\n")
		for _, seq := range schema.Sequences {
			renderSequence(sb, seq)
		}
		sb.WriteString("\n")
	}

	if len(schema.Triggers) > 0 {
		sb.WriteString("### Triggers\n\n")
		for _, trig := range schema.Triggers {
			renderTrigger(sb, trig)
		}
		sb.WriteString("\n")
	}

	if len(schema.Functions) > 0 {
		sb.WriteString("### Functions\n\n")
		for _, fn := range schema.Functions {
			renderFunction(sb, fn)
		}
		sb.WriteString("\n")
	}

	if len(schema.Types) > 0 {
		sb.WriteString("### Custom Types\n\n")
		for _, t := range schema.Types {
			renderType(sb, t)
		}
		sb.WriteString("\n")
	}
}

func renderTable(sb *strings.Builder, table pg.Table) {
	fmt.Fprintf(sb, "#### %s\n\n", table.Name)
	sb.WriteString("| Column | Type | Constraints |\n")
	sb.WriteString("|--------|------|-------------|\n")

	for _, col := range table.Columns {
		constraints := buildConstraints(col)
		fmt.Fprintf(sb, "| %s | %s | %s |\n", col.Name, col.Type, constraints)
	}

	if len(table.Indexes) > 0 {
		sb.WriteString("\n**Indexes:** ")
		var idxStrs []string
		for _, idx := range table.Indexes {
			idxStr := fmt.Sprintf("%s (%s", idx.Name, strings.Join(idx.Columns, ", "))
			if idx.IsPrimary {
				idxStr += ", PK"
			} else if idx.IsUnique {
				idxStr += ", UNIQUE"
			}
			idxStr += ")"
			idxStrs = append(idxStrs, idxStr)
		}
		sb.WriteString(strings.Join(idxStrs, ", "))
		sb.WriteString("\n")
	}

	sb.WriteString("\n")
}

func renderView(sb *strings.Builder, view pg.View) {
	fmt.Fprintf(sb, "#### %s\n\n", view.Name)
	sb.WriteString("| Column | Type |\n")
	sb.WriteString("|--------|------|\n")

	for _, col := range view.Columns {
		fmt.Fprintf(sb, "| %s | %s |\n", col.Name, col.Type)
	}

	sb.WriteString("\n")
}

func renderMaterializedView(sb *strings.Builder, mv pg.MaterializedView) {
	fmt.Fprintf(sb, "#### %s\n\n", mv.Name)
	sb.WriteString("| Column | Type |\n")
	sb.WriteString("|--------|------|\n")

	for _, col := range mv.Columns {
		fmt.Fprintf(sb, "| %s | %s |\n", col.Name, col.Type)
	}

	sb.WriteString("\n")
}

func renderSequence(sb *strings.Builder, seq pg.Sequence) {
	cycle := ""
	if seq.Cycle {
		cycle = ", CYCLE"
	}
	fmt.Fprintf(sb, "- `%s` (%s): start=%d, inc=%d, range=[%d..%d]%s\n",
		seq.Name, seq.DataType, seq.Start, seq.Increment, seq.Min, seq.Max, cycle)
}

func renderTrigger(sb *strings.Builder, trig pg.Trigger) {
	fmt.Fprintf(sb, "- `%s` on `%s`: %s %s → %s()\n",
		trig.Name, trig.Table, trig.Timing, trig.Event, trig.Function)
}

func renderFunction(sb *strings.Builder, fn pg.Function) {
	if fn.Arguments == "" {
		fmt.Fprintf(sb, "- `%s() → %s`\n", fn.Name, fn.ReturnType)
	} else {
		fmt.Fprintf(sb, "- `%s(%s) → %s`\n", fn.Name, fn.Arguments, fn.ReturnType)
	}
}

func renderType(sb *strings.Builder, t pg.CustomType) {
	if t.Kind == "enum" {
		var quoted []string
		for _, v := range t.Values {
			quoted = append(quoted, fmt.Sprintf("'%s'", v))
		}
		fmt.Fprintf(sb, "- `%s`: %s\n", t.Name, strings.Join(quoted, ", "))
	} else {
		fmt.Fprintf(sb, "- `%s` (composite): %s\n", t.Name, strings.Join(t.Values, ", "))
	}
}

func buildConstraints(col pg.Column) string {
	var parts []string

	if col.IsPK {
		parts = append(parts, "PK")
	}
	if !col.Nullable {
		parts = append(parts, "NOT NULL")
	}
	if col.IsUnique && !col.IsPK {
		parts = append(parts, "UNIQUE")
	}
	if col.FKRef != "" {
		parts = append(parts, fmt.Sprintf("FK→%s", col.FKRef))
	}
	if col.Default != "" {
		parts = append(parts, fmt.Sprintf("DEFAULT %s", col.Default))
	}

	return strings.Join(parts, ", ")
}

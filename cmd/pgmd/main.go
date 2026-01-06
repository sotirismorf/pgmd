package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/sotirismorf/pgmd/internal/markdown"
	"github.com/sotirismorf/pgmd/internal/pg"
)

func main() {
	uri := flag.String("uri", "", "PostgreSQL connection URI (required)")
	schemas := flag.String("schemas", "public", "Comma-separated schema names")
	flag.Parse()

	if *uri == "" {
		fmt.Fprintln(os.Stderr, "Error: -uri flag is required")
		fmt.Fprintln(os.Stderr, "Usage: pgmd -uri \"postgres://user:pass@host/db\" -schemas \"public,auth\"")
		os.Exit(1)
	}

	ctx := context.Background()

	conn, err := pgx.Connect(ctx, *uri)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error connecting to database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(ctx)

	schemaList := pg.ParseSchemas(*schemas)
	if len(schemaList) == 0 {
		fmt.Fprintln(os.Stderr, "Error: no schemas specified")
		os.Exit(1)
	}

	schemaInfos, err := pg.FetchSchemas(ctx, conn, schemaList)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching schema info: %v\n", err)
		os.Exit(1)
	}

	output := markdown.Render(schemaInfos)
	fmt.Print(output)
}

# pgmd

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go](https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go)](https://go.dev/)

PostgreSQL schema to Markdown. Generate LLM-friendly documentation of your database schema.

## Features

- Tables with columns, types, constraints (PK, FK, NOT NULL, UNIQUE, DEFAULT)
- Indexes
- Views and Materialized Views
- Sequences
- Triggers
- User-defined functions
- Custom types (enums, composites)

## Installation

```bash
go install github.com/sotirismorf/pgmd/cmd/pgmd@latest
```

Or build from source:

```bash
git clone https://github.com/sotirismorf/pgmd.git
cd pgmd
make build
```

## Usage

```bash
pgmd -uri "postgres://user:pass@localhost:5432/dbname" -schemas "public"
```

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-uri` | (required) | PostgreSQL connection URI |
| `-schemas` | `public` | Comma-separated list of schemas |

### Examples

Single schema:
```bash
pgmd -uri "postgres://localhost/mydb" -schemas "public"
```

Multiple schemas:
```bash
pgmd -uri "postgres://localhost/mydb" -schemas "public,auth,api"
```

Save to file:
```bash
pgmd -uri "postgres://localhost/mydb" > schema.md
```

## Output Format

```markdown
# Database Schema Documentation

## Schema: public

### Tables

#### users
| Column | Type | Constraints |
|--------|------|-------------|
| id | uuid | PK, NOT NULL |
| email | text | NOT NULL, UNIQUE |
| org_id | uuid | FK→orgs.id |

**Indexes:** users_pkey (id, PK), idx_users_email (email, UNIQUE)

### Views

#### active_users
| Column | Type |
|--------|------|
| id | uuid |
| email | text |

### Sequences

- `users_id_seq` (bigint): start=1, inc=1, range=[1..9223372036854775807]

### Triggers

- `update_timestamp` on `users`: BEFORE UPDATE → set_updated_at()

### Functions

- `get_user(id uuid) → users`

### Custom Types

- `status`: 'pending', 'active', 'archived'
```

## Use Cases

- Feed database context to LLM agents
- Generate documentation for your database
- Schema review and auditing
- Onboarding new developers

## License

MIT

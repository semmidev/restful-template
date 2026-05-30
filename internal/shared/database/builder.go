package database

import sq "github.com/Masterminds/squirrel"

// QB is the shared Squirrel statement builder pre-configured with PostgreSQL's
// dollar-sign placeholder format. Use this in all repository implementations
// to avoid re-declaring the builder in each package.
var QB = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

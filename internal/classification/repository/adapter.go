package repository

import (
	"context"

	"kyb-platform/internal/database"
	postgrest "github.com/supabase-community/postgrest-go"
)

// SupabaseClientAdapter adapts the real Supabase client to our interface
type SupabaseClientAdapter struct {
	client *database.SupabaseClient
}

// NewSupabaseClientAdapter creates a new adapter for the real Supabase client
func NewSupabaseClientAdapter(client *database.SupabaseClient) SupabaseClientInterface {
	return &SupabaseClientAdapter{client: client}
}

// Connect establishes a connection to Supabase
func (a *SupabaseClientAdapter) Connect(ctx context.Context) error {
	return a.client.Connect(ctx)
}

// Close closes the Supabase connection
func (a *SupabaseClientAdapter) Close() error {
	return a.client.Close()
}

// Ping checks the connection to Supabase
func (a *SupabaseClientAdapter) Ping(ctx context.Context) error {
	return a.client.Ping(ctx)
}

// GetClient returns the underlying Supabase client
func (a *SupabaseClientAdapter) GetClient() interface{} {
	return a.client.GetClient()
}

// GetPostgrestClient returns the PostgREST client for direct database operations
func (a *SupabaseClientAdapter) GetPostgrestClient() PostgrestClientInterface {
	return &PostgrestClientAdapter{client: a.client.GetPostgrestClient()}
}

// PostgrestClientAdapter adapts the real PostgREST client to our interface
type PostgrestClientAdapter struct {
	client interface{} // This will be *postgrest.Client
}

// From starts a query on a table
func (a *PostgrestClientAdapter) From(table string) PostgrestQueryInterface {
	// Use type assertion to get the real PostgREST client
	if client, ok := a.client.(*postgrest.Client); ok {
		return &RealPostgrestQuery{client: client, table: table}
	}
	// Fallback to basic implementation for tests
	return &BasicPostgrestQuery{}
}

// RealPostgrestQuery provides a real implementation using the actual PostgREST client
type RealPostgrestQuery struct {
	client *postgrest.Client
	table  string
	query  interface{} // This will hold the chained query
}

func (r *RealPostgrestQuery) Select(columns, count string, head bool) PostgrestQueryInterface {
	r.query = r.client.From(r.table).Select(columns, count, head)
	return r
}

func (r *RealPostgrestQuery) Eq(column, value string) PostgrestQueryInterface {
	if r.query != nil {
		// Use reflection to call the method on the chained query
		if q, ok := r.query.(interface {
			Eq(string, string) interface{}
		}); ok {
			r.query = q.Eq(column, value)
		}
	}
	return r
}

func (r *RealPostgrestQuery) Ilike(column, value string) PostgrestQueryInterface {
	if r.query != nil {
		// Use reflection to call the method on the chained query
		if q, ok := r.query.(interface {
			Ilike(string, string) interface{}
		}); ok {
			r.query = q.Ilike(column, value)
		}
	}
	return r
}

func (r *RealPostgrestQuery) In(column string, values ...string) PostgrestQueryInterface {
	if r.query != nil {
		// Use reflection to call the method on the chained query
		if q, ok := r.query.(interface {
			In(string, ...string) interface{}
		}); ok {
			r.query = q.In(column, values...)
		}
	}
	return r
}

func (r *RealPostgrestQuery) Order(column string, ascending *map[string]string) PostgrestQueryInterface {
	if r.query != nil {
		// Use reflection to call the method on the chained query
		if q, ok := r.query.(interface {
			Order(string, *map[string]string) interface{}
		}); ok {
			r.query = q.Order(column, ascending)
		}
	}
	return r
}

func (r *RealPostgrestQuery) Limit(count int, foreignTable string) PostgrestQueryInterface {
	if r.query != nil {
		// Use reflection to call the method on the chained query
		if q, ok := r.query.(interface{ Limit(int, string) interface{} }); ok {
			r.query = q.Limit(count, foreignTable)
		}
	}
	return r
}

func (r *RealPostgrestQuery) Single() PostgrestQueryInterface {
	if r.query != nil {
		// Use reflection to call the method on the chained query
		if q, ok := r.query.(interface{ Single() interface{} }); ok {
			r.query = q.Single()
		}
	}
	return r
}

func (r *RealPostgrestQuery) Execute() ([]byte, string, error) {
	if r.query != nil {
		// Use reflection to call the Execute method on the chained query
		if q, ok := r.query.(interface {
			Execute() ([]byte, string, error)
		}); ok {
			return q.Execute()
		}
	}
	return []byte{}, "", nil
}

// BasicPostgrestQuery provides a basic implementation for the adapter
type BasicPostgrestQuery struct{}

func (b *BasicPostgrestQuery) Select(columns, count string, head bool) PostgrestQueryInterface {
	return b
}
func (b *BasicPostgrestQuery) Eq(column, value string) PostgrestQueryInterface            { return b }
func (b *BasicPostgrestQuery) Ilike(column, value string) PostgrestQueryInterface         { return b }
func (b *BasicPostgrestQuery) In(column string, values ...string) PostgrestQueryInterface { return b }
func (b *BasicPostgrestQuery) Order(column string, ascending *map[string]string) PostgrestQueryInterface {
	return b
}
func (b *BasicPostgrestQuery) Limit(count int, foreignTable string) PostgrestQueryInterface { return b }
func (b *BasicPostgrestQuery) Single() PostgrestQueryInterface                              { return b }
func (b *BasicPostgrestQuery) Execute() ([]byte, string, error)                             { return []byte{}, "", nil }

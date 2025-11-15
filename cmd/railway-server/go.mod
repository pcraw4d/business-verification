module kyb-platform/cmd/railway-server

go 1.22

require (
	github.com/lib/pq v1.10.9
	github.com/supabase-community/postgrest-go v0.0.11
	go.uber.org/zap v1.27.0
	kyb-redis-optimization v0.0.0
)

replace kyb-redis-optimization => ../../pkg/redis-optimization

require (
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/redis/go-redis/v9 v9.3.0 // indirect
	go.uber.org/multierr v1.10.0 // indirect
)

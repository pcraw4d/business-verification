module kyb-platform/cmd/railway-server

go 1.22

require (
	github.com/supabase/postgrest-go v0.0.7
	kyb-redis-optimization v0.0.0
)

replace kyb-redis-optimization => ../../pkg/redis-optimization

require (
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/jarcoal/httpmock v1.3.1 // indirect
	github.com/redis/go-redis/v9 v9.3.0 // indirect
	github.com/stretchr/testify v1.8.4 // indirect
)

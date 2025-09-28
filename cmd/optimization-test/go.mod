module kyb-optimization-test

go 1.22

require (
	kyb-api-optimization v0.0.0
	kyb-database-optimization v0.0.0
	kyb-monitoring-optimization v0.0.0
	kyb-redis-optimization v0.0.0
)

require (
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/redis/go-redis/v9 v9.3.0 // indirect
	github.com/supabase/postgrest-go v0.0.7 // indirect
)

replace kyb-api-optimization => ../../pkg/api-optimization

replace kyb-database-optimization => ../../pkg/database-optimization

replace kyb-monitoring-optimization => ../../pkg/monitoring-optimization

replace kyb-redis-optimization => ../../pkg/redis-optimization

module kyb-advanced-gateway

go 1.22

require (
	github.com/gorilla/mux v1.8.1
	github.com/supabase/postgrest-go v0.0.7
	kyb-advanced-analytics v0.0.0
	kyb-multi-tenant v0.0.0
)

replace kyb-advanced-analytics => ../../pkg/advanced-analytics

replace kyb-multi-tenant => ../../pkg/multi-tenant

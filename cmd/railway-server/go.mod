module kyb-platform/cmd/railway-server

go 1.22

require (
	kyb-platform/pkg/cache v0.0.0
	kyb-platform/pkg/monitoring v0.0.0
	kyb-platform/pkg/performance v0.0.0
	kyb-platform/pkg/security v0.0.0
	kyb-platform/pkg/analytics v0.0.0
	kyb-platform/pkg/api v0.0.0
	github.com/gorilla/mux v1.8.1
	github.com/supabase-community/supabase-go v0.0.1
	go.uber.org/zap v1.27.0
)

replace kyb-platform/pkg/cache => ../../pkg/cache

replace kyb-platform/pkg/monitoring => ../../pkg/monitoring

replace kyb-platform/pkg/performance => ../../pkg/performance

replace kyb-platform/pkg/security => ../../pkg/security

replace kyb-platform/pkg/analytics => ../../pkg/analytics

replace kyb-platform/pkg/api => ../../pkg/api

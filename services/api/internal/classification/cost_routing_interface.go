package classification

// CostRoutingMethodRegistry defines the interface for method registry used by cost-based router
type CostRoutingMethodRegistry interface {
	GetAllMethods() []ClassificationMethod
	GetMethod(name string) (ClassificationMethod, error)
	RegisterMethod(method ClassificationMethod, config MethodConfig) error
	UnregisterMethod(name string) error
	UpdateMethodConfig(name string, config MethodConfig) error
}

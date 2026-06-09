package gemini

import (
	"live-api/internal/mcp"
	"live-api/internal/mcp/registry"

	"google.golang.org/genai"
)

type ToolRouter struct {
	Discovery *registry.Discovery
}

func (r ToolRouter) Declarations(useCase registry.UseCase) []*genai.FunctionDeclaration {
	if r.Discovery == nil {
		return AIVisibleCoreTools()
	}
	return CoreFunctionDeclarations(r.Discovery.GetToolsForUseCase(useCase))
}

func RegistryRouter(reg *mcp.ToolRegistry) ToolRouter {
	return ToolRouter{Discovery: registry.NewDiscovery(reg)}
}

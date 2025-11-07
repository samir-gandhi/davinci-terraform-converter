1. The Provider Interface (The Core Abstraction)

This is the most important addition. We define a generic interface that any service client (PingOne, Okta, Auth0, etc.) must implement to be "discoverable."

- Provider: An interface that represents a single data source (like the PingOne API). It's responsible for its own authentication and for fetching its specific resources.

- Resource: A generic struct to hold the data for a single discovered item, crucial for passing information between stages.

```Go

// Provider defines the contract for any data source.
// Each "provider" (PingOne, Okta, etc.) will have a concrete
// struct that implements this interface.
type Provider interface {
    // Name returns a unique identifier, e.g., "pingone" or "auth0".
    Name() string
    
    // DiscoverResources fetches all resources from its API
    // and returns them as a slice of generic Resource structs.
    DiscoverResources(ctx context.Context) ([]*Resource, error)

    // Configure sets up the provider's API client using generic auth config.
    Configure(authConfig map[string]string) error
}

// Resource is the common data structure passed between pipeline stages.
// It represents a single discovered API object (e.g., one application).
type Resource struct {
    ProviderName string // "pingone"
    Type         string // "application"
    ID           string
    Data         json.RawMessage // The raw JSON from the API
}
```

2. The Discovery Engine

The DiscoveryEngine is now just a manager for a list of Provider interfaces. It no longer knows or cares about any specific SDK.

- DiscoveryEngine: Holds a registry of all active providers.

- ResourceCache: This is now a simple map[string]*Resource, holding all resources from all providers, keyed by a globally unique ID (e.g., pingone_application_12345).

```Go

// DiscoveryEngine manages running one or more providers.
type DiscoveryEngine struct {
    Providers []Provider // A list of configured providers to run
}

// ResourceCache holds the raw, unstructured resources from all providers.
type ResourceCache map[string]*Resource
```

3. The Dependency Mapper

This component's structs remain largely the same, but they operate on the generic Resource struct, using its ProviderName and Type fields to apply the correct rules.

- DependencyMapper: Holds the mapping rules.

- ResourceGraph: The DAG.

- ResourceNode: A node in the graph. It's essentially just our Resource struct, now placed within the graph.


```Go

// DependencyMapper builds the resource graph.
type DependencyMapper struct {
    // Rules are now likely a map, keyed by provider name,
    // e.g., map[string]*ProviderRules
    Rules *MappingSchema
}

// ResourceNode represents a single resource (a node in the graph).
// It can simply embed the generic Resource struct.
type ResourceNode struct {
    *Resource
    // Dependencies (edges) would be managed by the graph struct.
}
```

4. The HCL Generation Engine

This engine must also be provider-aware. It will hold a collection of mapping schemas and use the Resource.ProviderName to select the correct one for translation.

- HCLGenerationEngine: Holds a registry of mapping schemas.

- SchemaRegistry: A simple map that keys a ProviderName (e.g., "pingone") to its specific MappingSchema.

- MappingSchema: This is the same struct as before (defining API-to-HCL rules), but now we'll have one per provider.

```Go

// SchemaRegistry holds the HCL mapping rules for all known providers.
type SchemaRegistry map[string]*MappingSchema // key: provider name

// HCLGenerationEngine translates the sorted resource graph into HCL.
type HCLGenerationEngine struct {
    // It holds the registry of *all* possible schemas.
    SchemaRegistry SchemaRegistry
}

// MappingSchema (Same as before)
// Defines the translation logic from API responses to HCL
// for a *single* provider.
type MappingSchema struct {
    Resources map[string]ResourceMapping // key: resource type
    // ...
}
```

5. The Orchestrator & CLI

The Orchestrator is responsible for initializing the correct Provider implementations based on user config and injecting them into the DiscoveryEngine.

```Go

// Orchestrator manages the end-to-end export pipeline.
type Orchestrator struct {
    Config          *CLIConfig
    DiscoveryEngine *DiscoveryEngine
    DependencyMapper  *DependencyMapper
    HCLGenEngine    *HCLGenerationEngine
    
    // Holds all *available* provider factories
    ProviderFactory map[string]func() Provider
}

// CLIConfig holds configuration provided by the user via flags.
type CLIConfig struct {
    OutputDir   string
    // User might specify which providers to run
    Providers   []string // e.g., ["pingone", "okta"]
    // Config would be nested, e.g.,
    // Auth:
    //   pingone: { client_id: "...", ... }
    //   okta:    { org_url: "...", ... }
}
```
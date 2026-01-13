package provider

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the provider.Provider interface.
var _ provider.Provider = &DemoAppProvider{}

// DemoAppClient is the client that resources will use to talk to the demo-app API.
// We create this in Configure() and pass it to all resources.
type DemoAppClient struct {
	// HTTPClient is the underlying HTTP client
	HTTPClient *http.Client

	// Endpoint is the base URL of the demo-app API (e.g., "http://localhost:8080")
	Endpoint string
}

// DemoAppProvider defines the provider implementation.
type DemoAppProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and run locally.
	version string
}

// DemoAppProviderModel describes the provider data model.
// This maps to the provider block in HCL.
type DemoAppProviderModel struct {
	Endpoint types.String `tfsdk:"endpoint"`
}

// New is a helper function to simplify provider server construction.
// This is what main.go calls: provider.New(version)
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &DemoAppProvider{
			version: version,
		}
	}
}

// Metadata returns the provider type name.
func (p *DemoAppProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "demoapp"
}

// Schema defines the provider-level schema.
// This is what users configure in their provider block.
func (p *DemoAppProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Interact with the Demo App API.",
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				Description: "The endpoint URL of the Demo App API (e.g., http://localhost:8080). Can also be set via DEMOAPP_ENDPOINT environment variable.",
				Optional:    true,
			},
		},
	}
}

// Configure prepares the API client for data sources and resources.
// This is called once when Terraform initializes the provider.
func (p *DemoAppProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config DemoAppProviderModel

	// Read the provider configuration from HCL into our model
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Determine the endpoint: HCL config takes priority, then environment variable
	// This is a common pattern - let users set via provider block OR environment
	endpoint := os.Getenv("DEMOAPP_ENDPOINT")

	// If endpoint is set in HCL config, use that instead
	if !config.Endpoint.IsNull() {
		endpoint = config.Endpoint.ValueString()
	}

	// If we still don't have an endpoint, that's an error
	if endpoint == "" {
		resp.Diagnostics.AddError(
			"Missing Demo App Endpoint",
			"The provider cannot create the Demo App API client because the endpoint is missing. "+
				"Set the endpoint in the provider configuration or via the DEMOAPP_ENDPOINT environment variable.",
		)
		return
	}

	// Create the HTTP client with reasonable defaults
	// 30 second timeout prevents hanging forever on network issues
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Create our client wrapper
	client := &DemoAppClient{
		HTTPClient: httpClient,
		Endpoint:   endpoint,
	}

	// Pass the client to all resources and data sources
	// When a resource's Configure() method is called, it receives this via req.ProviderData
	resp.DataSourceData = client
	resp.ResourceData = client
}

// DataSources defines the data sources implemented in the provider.
func (p *DemoAppProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		// We'll add data sources here later
	}
}

// Resources defines the resources implemented in the provider.
func (p *DemoAppProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewItemResource,
		NewDisplayResource,
	}
}

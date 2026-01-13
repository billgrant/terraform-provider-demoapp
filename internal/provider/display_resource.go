package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Compile-time interface check
var _ resource.Resource = &DisplayResource{}

// DisplayResource manages the display panel content.
// Unlike items, there's only ONE display — it's a singleton.
// Each POST replaces the previous content entirely.
type DisplayResource struct {
	client *DemoAppClient
}

// DisplayResourceModel maps to the Terraform configuration.
// Just one attribute: the JSON data to display.
type DisplayResourceModel struct {
	// ID is required by Terraform but meaningless for a singleton
	// We'll just use a fixed value like "display"
	ID types.String `tfsdk:"id"`

	// Data is the JSON content to show in the display panel
	// User passes a JSON string, we POST it to the API
	Data types.String `tfsdk:"data"`
}

// NewDisplayResource is the factory function.
func NewDisplayResource() resource.Resource {
	return &DisplayResource{}
}

// Metadata sets the resource type name: demoapp_display
func (r *DisplayResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_display"
}

// Schema defines what users can configure.
func (r *DisplayResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages the display panel content in Demo App. Posts arbitrary JSON data that the frontend renders.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Placeholder ID (always 'display' since there's only one display panel).",
				Computed:    true,
			},

			"data": schema.StringAttribute{
				Description: "JSON string to display. Use jsonencode() to convert HCL to JSON.",
				Required:    true,
			},
		},
	}
}

// Configure receives the provider's HTTP client.
func (r *DisplayResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*DemoAppClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *DemoAppClient, got: %T", req.ProviderData),
		)
		return
	}

	r.client = client
}

// Create posts the JSON data to the display endpoint.
func (r *DisplayResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan DisplayResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate that the data is valid JSON
	if !json.Valid([]byte(plan.Data.ValueString())) {
		resp.Diagnostics.AddError(
			"Invalid JSON",
			"The 'data' attribute must be valid JSON. Use jsonencode() to convert HCL maps to JSON.",
		)
		return
	}

	// POST the JSON to /api/display
	url := r.client.Endpoint + "/api/display"
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBufferString(plan.Data.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Display",
			"Could not create HTTP request: "+err.Error(),
		)
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")

	httpResp, err := r.client.HTTPClient.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Display",
			"Could not send HTTP request: "+err.Error(),
		)
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK && httpResp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(httpResp.Body)
		resp.Diagnostics.AddError(
			"Error Creating Display",
			fmt.Sprintf("API returned status %d: %s", httpResp.StatusCode, string(body)),
		)
		return
	}

	// Set the ID (fixed value since display is a singleton)
	plan.ID = types.StringValue("display")

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read fetches the current display content.
func (r *DisplayResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state DisplayResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// GET the current display content
	url := r.client.Endpoint + "/api/display"
	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Display",
			"Could not create HTTP request: "+err.Error(),
		)
		return
	}

	httpResp, err := r.client.HTTPClient.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Display",
			"Could not send HTTP request: "+err.Error(),
		)
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(httpResp.Body)
		resp.Diagnostics.AddError(
			"Error Reading Display",
			fmt.Sprintf("API returned status %d: %s", httpResp.StatusCode, string(body)),
		)
		return
	}

	// Read the response body as raw JSON
	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Display",
			"Could not read response body: "+err.Error(),
		)
		return
	}

	// Update state with current data from API
	state.Data = types.StringValue(string(body))
	state.ID = types.StringValue("display")

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update is the same as Create for display — just POST new content.
func (r *DisplayResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan DisplayResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate JSON
	if !json.Valid([]byte(plan.Data.ValueString())) {
		resp.Diagnostics.AddError(
			"Invalid JSON",
			"The 'data' attribute must be valid JSON. Use jsonencode() to convert HCL maps to JSON.",
		)
		return
	}

	// POST the new content (same as Create)
	url := r.client.Endpoint + "/api/display"
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBufferString(plan.Data.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Display",
			"Could not create HTTP request: "+err.Error(),
		)
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")

	httpResp, err := r.client.HTTPClient.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Display",
			"Could not send HTTP request: "+err.Error(),
		)
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK && httpResp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(httpResp.Body)
		resp.Diagnostics.AddError(
			"Error Updating Display",
			fmt.Sprintf("API returned status %d: %s", httpResp.StatusCode, string(body)),
		)
		return
	}

	plan.ID = types.StringValue("display")
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete clears the display by posting empty JSON.
func (r *DisplayResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// For display, "delete" means clear it — post empty object
	url := r.client.Endpoint + "/api/display"
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBufferString("{}"))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Display",
			"Could not create HTTP request: "+err.Error(),
		)
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")

	httpResp, err := r.client.HTTPClient.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Display",
			"Could not send HTTP request: "+err.Error(),
		)
		return
	}
	defer httpResp.Body.Close()

	// We don't really care about the response for delete
	// Just let Terraform remove it from state
}

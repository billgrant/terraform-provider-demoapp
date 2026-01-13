package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Compile-time check: does ItemResource implement resource.Resource?
var _ resource.Resource = &ItemResource{}

// ItemResource defines the resource implementation.
type ItemResource struct {
	// client is the configured DemoAppClient from the provider
	client *DemoAppClient
}

// ItemResourceModel describes the resource data model.
// This struct maps to both:
//   - The HCL the user writes (terraform configuration)
//   - The state file (what Terraform remembers)
type ItemResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
}

// itemAPIModel represents the JSON structure from the demo-app API.
// This is separate from ItemResourceModel because:
//   - API uses int for ID, Terraform uses string
//   - API might have fields we don't expose to Terraform
//   - Keeps API concerns separate from Terraform concerns
type itemAPIModel struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// NewItemResource is the factory function that creates instances of this resource.
func NewItemResource() resource.Resource {
	return &ItemResource{}
}

// Metadata sets the resource type name.
func (r *ItemResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_item"
}

// Schema defines the structure of the resource.
func (r *ItemResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an item in the Demo App.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the item.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			"name": schema.StringAttribute{
				Description: "The name of the item.",
				Required:    true,
			},

			"description": schema.StringAttribute{
				Description: "A description of the item.",
				Optional:    true,
			},
		},
	}
}

// Configure receives the provider's configured client.
func (r *ItemResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Provider hasn't been configured yet
	if req.ProviderData == nil {
		return
	}

	// Type assertion: convert the generic interface{} to our specific type
	// This is Go's way of saying "I know this is a *DemoAppClient, trust me"
	// The ", ok" pattern checks if the assertion succeeded
	client, ok := req.ProviderData.(*DemoAppClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *DemoAppClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

// Create makes a POST request to create a new item.
func (r *ItemResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// 1. Read the planned values from Terraform configuration
	var plan ItemResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// 2. Build the request body
	// We convert from Terraform types to plain Go types for JSON encoding
	requestBody := itemAPIModel{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Item",
			"Could not marshal request body: "+err.Error(),
		)
		return
	}

	// 3. Make the HTTP request
	url := r.client.Endpoint + "/api/items"
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Item",
			"Could not create HTTP request: "+err.Error(),
		)
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")

	httpResp, err := r.client.HTTPClient.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Item",
			"Could not send HTTP request: "+err.Error(),
		)
		return
	}
	defer httpResp.Body.Close()

	// 4. Check for errors
	if httpResp.StatusCode != http.StatusCreated && httpResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(httpResp.Body)
		resp.Diagnostics.AddError(
			"Error Creating Item",
			fmt.Sprintf("API returned status %d: %s", httpResp.StatusCode, string(body)),
		)
		return
	}

	// 5. Parse the response
	var apiResponse itemAPIModel
	if err := json.NewDecoder(httpResp.Body).Decode(&apiResponse); err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Item",
			"Could not parse API response: "+err.Error(),
		)
		return
	}

	// 6. Update the plan with values from the API response
	// The API gives us the ID, which we need to store in state
	plan.ID = types.StringValue(strconv.Itoa(apiResponse.ID))
	plan.Name = types.StringValue(apiResponse.Name)
	plan.Description = types.StringValue(apiResponse.Description)

	// 7. Save the state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read fetches the current state from the API.
func (r *ItemResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// 1. Read the current state
	var state ItemResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// 2. Make the HTTP request
	url := r.client.Endpoint + "/api/items/" + state.ID.ValueString()
	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Item",
			"Could not create HTTP request: "+err.Error(),
		)
		return
	}

	httpResp, err := r.client.HTTPClient.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Item",
			"Could not send HTTP request: "+err.Error(),
		)
		return
	}
	defer httpResp.Body.Close()

	// 3. Handle 404 - resource was deleted outside Terraform
	if httpResp.StatusCode == http.StatusNotFound {
		// Tell Terraform the resource no longer exists
		// This will show as "will be created" in the next plan
		resp.State.RemoveResource(ctx)
		return
	}

	// 4. Check for other errors
	if httpResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(httpResp.Body)
		resp.Diagnostics.AddError(
			"Error Reading Item",
			fmt.Sprintf("API returned status %d: %s", httpResp.StatusCode, string(body)),
		)
		return
	}

	// 5. Parse the response
	var apiResponse itemAPIModel
	if err := json.NewDecoder(httpResp.Body).Decode(&apiResponse); err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Item",
			"Could not parse API response: "+err.Error(),
		)
		return
	}

	// 6. Update state with current values from API
	state.ID = types.StringValue(strconv.Itoa(apiResponse.ID))
	state.Name = types.StringValue(apiResponse.Name)
	state.Description = types.StringValue(apiResponse.Description)

	// 7. Save the refreshed state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update makes a PUT request to update an existing item.
func (r *ItemResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// 1. Read the planned new values
	var plan ItemResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// 2. Build the request body
	requestBody := itemAPIModel{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Item",
			"Could not marshal request body: "+err.Error(),
		)
		return
	}

	// 3. Make the HTTP request
	url := r.client.Endpoint + "/api/items/" + plan.ID.ValueString()
	httpReq, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Item",
			"Could not create HTTP request: "+err.Error(),
		)
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")

	httpResp, err := r.client.HTTPClient.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Item",
			"Could not send HTTP request: "+err.Error(),
		)
		return
	}
	defer httpResp.Body.Close()

	// 4. Check for errors
	if httpResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(httpResp.Body)
		resp.Diagnostics.AddError(
			"Error Updating Item",
			fmt.Sprintf("API returned status %d: %s", httpResp.StatusCode, string(body)),
		)
		return
	}

	// 5. Parse the response
	var apiResponse itemAPIModel
	if err := json.NewDecoder(httpResp.Body).Decode(&apiResponse); err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Item",
			"Could not parse API response: "+err.Error(),
		)
		return
	}

	// 6. Update plan with values from API response
	plan.ID = types.StringValue(strconv.Itoa(apiResponse.ID))
	plan.Name = types.StringValue(apiResponse.Name)
	plan.Description = types.StringValue(apiResponse.Description)

	// 7. Save the updated state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete makes a DELETE request to remove an item.
func (r *ItemResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// 1. Read the current state to get the ID
	var state ItemResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// 2. Make the HTTP request
	url := r.client.Endpoint + "/api/items/" + state.ID.ValueString()
	httpReq, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Item",
			"Could not create HTTP request: "+err.Error(),
		)
		return
	}

	httpResp, err := r.client.HTTPClient.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Item",
			"Could not send HTTP request: "+err.Error(),
		)
		return
	}
	defer httpResp.Body.Close()

	// 3. Check for errors (404 is okay - already deleted)
	if httpResp.StatusCode != http.StatusOK && httpResp.StatusCode != http.StatusNoContent && httpResp.StatusCode != http.StatusNotFound {
		body, _ := io.ReadAll(httpResp.Body)
		resp.Diagnostics.AddError(
			"Error Deleting Item",
			fmt.Sprintf("API returned status %d: %s", httpResp.StatusCode, string(body)),
		)
		return
	}

	// 4. Terraform automatically removes from state after Delete returns successfully
}

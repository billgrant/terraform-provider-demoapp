# Terraform Provider DemoApp - Development Log

## Session Instructions (IMPORTANT)

**User's special instructions for Phase 5:**

> I have been a terraform practitioner for over 8 years, a tf sales engineer for 4. I have never built my own provider. I really want to dig in, line by line if needed I want to understand everything we code even more than normal. This is super interesting to me and I think it will be a valuable thing for me to learn.

**This means:**
- Go slow, explain every line of code
- Explain programming concepts that may be unfamiliar (factory functions, interfaces, etc.)
- Connect Terraform provider internals to existing TF practitioner knowledge
- This is a learning session, not a "get it done" session

---

## Session 1 - 2026-01-12

### Concepts Covered

Before writing any code, we discussed:

1. **Provider Architecture**
   - Terraform Core and Provider are separate processes
   - They communicate over gRPC (hidden from users)
   - Core is the orchestrator, Provider is the "driver" (like a printer driver)
   - Core handles state, plans, dependency graphs
   - Provider handles all API calls to the target system

2. **The Two SDKs**
   - SDKv2 (old, being phased out)
   - Plugin Framework (new, what we're using)

3. **Key Terraform Concepts Clarified**
   - Read() is only called for resources already in state
   - New resources go straight to Create() without checking if they exist
   - This is why `terraform import` exists
   - Provider makes ALL API calls, Core never talks to external systems

4. **Go Concepts Explained**
   - Factory functions: Functions that create and return objects (like constructors in other languages)
   - Blocking functions: `providerserver.Serve()` blocks forever, only returns on error
   - Interface checks: `var _ provider.Provider = &DemoAppProvider{}` is compile-time verification
   - `types.String` vs `string`: Framework types can represent null/unknown states

### Files Created

```
terraform-provider-demoapp/
├── main.go                           # Entry point, starts gRPC server
├── go.mod                            # Dependencies (terraform-plugin-framework)
├── go.sum                            # Dependency checksums
├── internal/
│   └── provider/
│       ├── provider.go               # Provider config (endpoint attribute)
│       └── item_resource.go          # demoapp_item resource (CRUD stubs)
└── examples/                         # (empty, for test configs later)
```

### Current State

**Completed:**
- main.go - fully implemented, explained line by line
- go.mod/go.sum - dependencies resolved
- provider.go - structure complete, Configure() has TODOs for HTTP client
- item_resource.go - structure complete with all CRUD method stubs

**In Progress:**
- item_resource.go needs actual HTTP client calls wired up

**Next Steps:**
1. Wire up HTTP client in provider.go Configure()
2. Implement actual HTTP calls in item_resource.go (Create, Read, Update, Delete)
3. Register ItemResource in provider.go's Resources() method
4. Create demoapp_display resource
5. Test locally with a running demo-app instance

### Code Walkthrough Status

Explained in detail:
- [x] main.go - all lines
- [x] provider.go - all lines
- [x] item_resource.go - structure and data model explained
- [ ] item_resource.go - HTTP implementation (next)

### Key Mental Models Established

- "Provider is like a printer driver - speaks the OS interface AND the printer's protocol"
- "Factory function returns a function that creates things - like a blueprint vs a house"
- "Core decides WHAT needs to happen, Provider knows HOW to make it happen"
- "Schema is metadata for Core, Provider does the actual API work"

---

## Session 2 - 2026-01-13

### Completed the Provider

Picked up where Session 1 left off. Implemented everything needed for a working provider.

### Go Concepts Explained

1. **`ok` pattern vs `err` pattern**
   - `err` is an error object with details
   - `ok` is a boolean for type assertions and map lookups
   - Both used for "did this work?" but different mechanisms

2. **`types.StringValue()` and `ValueString()`**
   - Terraform framework methods for converting between Terraform and Go types
   - `types.StringValue("foo")` — Go string → Terraform types.String
   - `plan.Name.ValueString()` — Terraform types.String → Go string

3. **`context.Context`**
   - Go's standard "kill switch" that gets passed around
   - Allows cancellation to propagate through call chains
   - Used everywhere for I/O operations (HTTP requests, DB queries)
   - When Terraform gets Ctrl+C, context signals provider to abort

4. **Why Go uses factory functions**
   - Go doesn't have constructors like Python/Java
   - `NewXxx()` functions fill that role
   - Terraform uses factory functions for deferred/repeated instantiation

### Implementation Completed

1. **HTTP Client in provider.go**
   - Created `DemoAppClient` struct
   - Endpoint from config or `DEMOAPP_ENDPOINT` env var
   - Passed to resources via `resp.ResourceData`

2. **item_resource.go CRUD**
   - Full Create/Read/Update/Delete implementation
   - Two models: `ItemResourceModel` (Terraform) and `itemAPIModel` (API)
   - 404 handling in Read() removes from state

3. **display_resource.go**
   - Singleton resource (one display panel)
   - Accepts arbitrary JSON via `jsonencode()`
   - Delete clears by posting empty `{}`

### Testing

All CRUD operations tested successfully:

| Operation | Result |
|-----------|--------|
| Create | 2 items + display created |
| Read | State refreshed correctly |
| Update | In-place update worked |
| Delete | Removed from API |
| Destroy | All resources cleaned up |

### Files Added

```
terraform-provider-demoapp/
├── .gitignore
├── .goreleaser.yml            # For registry releases
├── README.md                   # Full documentation
├── docs/
│   ├── index.md               # Provider docs (registry format)
│   └── resources/
│       ├── item.md            # demoapp_item docs
│       └── display.md         # demoapp_display docs
└── examples/
    └── main.tf                # Test configuration
```

### Registry Publishing Notes

To publish to registry.terraform.io:
1. Push to GitHub as public repo
2. Set up GPG key for signing
3. Configure GitHub Actions with GoReleaser
4. Register at registry.terraform.io
5. Create a release (tag + push)

### Vision

User's vision for the provider:
> "TF provider will be used more often than not when using the app. It would create a state in a stateless app. App crashed mid demo, re-run tf and you get your stuff back."

Terraform becomes the persistence layer for a stateless demo app.

---

## Known Issues

### SQLite Concurrency (Demo App Bug)

Demo App uses SQLite which doesn't handle concurrent writes well. When Terraform creates multiple resources in parallel, you may see:

```
Error: Error Creating Item
API returned status 500: {"error":"database error"}
```

**Workaround:** Use `-parallelism=1` to force sequential operations:

```bash
terraform apply -parallelism=1
terraform destroy -parallelism=1
```

**Fix:** Coming in Demo App Phase 7 — WAL mode + busy timeout will allow concurrent operations.

---

## Blog Post Note

Phase 4 (Docker) and Phase 5 (Provider) will be combined into one blog post since Phase 4 was short.

<!-- BEGIN:backend-go-rules -->
# Backend Go Development Rules

This file defines mandatory conventions for all Go backend development in this project.

## 1. Swagger Documentation Required

Every API handler MUST include `// swagger:route` annotations documenting:
- Route path and HTTP method
- Parameters (path, query, body)
- Response codes and types

Example:
```go
// ListNotifications returns all notifications for the authenticated user
//
// swagger:route GET /api/notifications notifications listNotifications
//
// Responses:
//   200: []NotificationResponse
//   401: ErrorResponse
func (h *Handler) ListNotifications(w http.ResponseWriter, r *http.Request) {
```

Do NOT merge the Swagger JSON manually. Update the swagger annotations in handler files and regenerate via `swagger generate spec`.

## 2. Strict File Separation (per package)

Each backend package MUST follow this 3-file structure:

| File | Purpose |
|------|---------|
| `types.go` | Struct definitions, request/response types, constants only — NO methods beyond simple constructors |
| `service.go` | Business logic, repository calls, validation — `Service` struct with methods |
| `handler.go` | HTTP handler methods — request parsing, response writing, delegation to service |

Rule: a single .go file MUST NOT contain both a type definition and business logic/HTTP handler code.

## 3. Service Layer Mandatory

- Handlers MUST only: parse request data, call service methods, write responses
- Service layer MUST contain all business logic, repository calls, and cross-cutting coordination
- Services are injected into handlers via constructor (`NewHandler(svc *Service)`)
- Services may depend on other services (e.g., notification service injected into message service)

## 4. Shared HTTP Utilities

Use `httputil.RespondJSON(w, status, data)` and `httputil.RespondError(w, status, message)` exclusively.
Do NOT write raw `json.NewEncoder(w).Encode(...)` or `w.Write(...)` in handlers.

## 5. Import Rules

- Internal packages are imported as `github.com/CpBruceMeena/sync/internal/<package>`
- Never import `internal/handler` packages from outside their intended scope
- Services import repositories; handlers import services — no circular dependencies
<!-- END:backend-go-rules -->

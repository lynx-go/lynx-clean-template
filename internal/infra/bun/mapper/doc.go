// Package mapper provides bidirectional conversion functions between
// domain entity types and Bun ORM model types.
//
// Naming convention:
//   - ToDomain*   converts a Bun model → domain entity
//   - FromDomain* converts a domain entity → Bun model  (for create/update)
//
// One file per aggregate (users.go, wallets.go, …).
// Repository files (bunrepo/) should delegate all struct conversions here
// so they only contain query logic.
package mapper

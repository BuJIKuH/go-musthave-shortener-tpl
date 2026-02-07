// Package noosexit contains an analyzer that forbids
// direct calls to os.Exit inside the main function
// of package main.
//
// Rationale:
//
// Calling os.Exit from main breaks:
//   - deferred calls
//   - graceful shutdown
//   - fx.Lifecycle and similar frameworks
//
// Instead, errors should be returned and handled
// at the application boundary.
package noosexit

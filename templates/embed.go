package templates

import "embed"

// FS contiene los 5 skills predefinidos.
//
//go:embed sales-closer content-engine lead-nurture life-os creator-stack
var FS embed.FS

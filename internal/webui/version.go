package webui

// Version is the asset version used to cache-bust CSS and JS URLs. It is set
// at build time with:
//
//	-ldflags "-X pizza-son/internal/webui.Version=<git short hash>"
//
// When it is left at "dev", assets are served with revalidation instead of a
// long-lived immutable cache.
var Version = "dev"

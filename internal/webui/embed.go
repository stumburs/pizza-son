package webui

import "embed"

// Files contains the static web UI served by the dashboard. Keeping it in the
// binary means deployments no longer need the static files alongside the
// executable.
//
//go:embed all:public
var Files embed.FS

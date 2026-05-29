package main

import (
	"fmt"
	"time"
)

// revision is set at build time via -ldflags "-X main.revision=<value>".
// In Nix builds this is the flake revision; in local builds it defaults to
// "dev-<timestamp>" so that each process start busts the asset cache.
var revision = "dev"

func init() {
	if revision == "dev" {
		revision = fmt.Sprintf("dev-%d", time.Now().UnixMilli())
	}
}

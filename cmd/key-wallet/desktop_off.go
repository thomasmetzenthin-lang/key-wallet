//go:build !desktop

package main

import "fmt"

func initDesktopFeatures() {
fmt.Println("[SYS] Mode: Headless Daemon (Minimal Footprint)")
}

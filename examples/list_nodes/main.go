// Example: list cluster nodes.
//
// Run with:
//
//	PBS_HOST=https://pbs.example.com:8007 \
//	PBS_TOKEN='PBSAPIToken=root@pam!auto:...' \
//	go run ./examples/list_nodes
package main

import (
	"context"
	"fmt"
	"os"

	pbs "github.com/client-api/pbs-go"
)

func main() {
	host := os.Getenv("PBS_HOST")
	if host == "" {
		host = "https://localhost:8007"
	}
	cfg := pbs.NewConfiguration()
	cfg.Servers = append(pbs.ServerConfigurations{}, pbs.ServerConfiguration{URL: host + "/api2/json"})
	cfg.DefaultHeader["Authorization"] = os.Getenv("PBS_TOKEN")

	client := pbs.NewAPIClient(cfg)
	resp, _, err := client.NodesAPI.NodesGetNodes(context.Background()).Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, "list nodes:", err)
		os.Exit(1)
	}
	// Non-PVE products: the upstream apidoc.js declares this endpoint
	// `returns: { type: null }`, so the generator emits `data` as an
	// untyped `interface{}` field. Print whatever came back.
	fmt.Printf("Response: %+v\n", resp.GetData())
}

// Example: open a terminal session against a QEMU VM.
//
// Run with:
//
//	PBS_HOST=https://pbs.example.com:8007 \
//	PBS_TOKEN='PBSAPIToken=root@pam!auto=...' \
//	PBS_NODE=orca PBS_VMID=100 \
//	go run ./examples/terminal
package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	pbs "github.com//"
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
	node := envOr("PBS_NODE", "pbs1")
	vmid64, _ := strconv.ParseInt(envOr("PBS_VMID", "100"), 10, 32)
	vmid := int32(vmid64)

	target := pbs.Target{Kind: pbs.TargetKindQemu, Node: node, Vmid: vmid}
	fmt.Printf("Opening terminal on %s:qemu/%d...\n", node, vmid)

	session, err := client.ConnectTerminal(context.Background(), target)
	if err != nil {
		fmt.Fprintln(os.Stderr, "connect:", err)
		os.Exit(1)
	}
	session.OnMessage = func(text string) { fmt.Print(text) }
	session.OnClose = func(err error) { fmt.Printf("\n[closed: %v]\n", err) }

	if err := session.Resize(120, 32); err != nil {
		fmt.Fprintln(os.Stderr, "resize:", err)
	}
	if err := session.Send("uname -a\n"); err != nil {
		fmt.Fprintln(os.Stderr, "send:", err)
	}
	time.Sleep(5 * time.Second)
	_ = session.Close()
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

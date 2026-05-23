// Example: resilient terminal session with auto-reconnect.
//
// Run with:
//
//	PBS_HOST=https://pbs.example.com:8007 \
//	PBS_TOKEN='PBSAPIToken=root@pam!auto=...' \
//	PBS_NODE=orca PBS_VMID=100 \
//	go run ./examples/resilient_terminal
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
	opts := pbs.RetryOptions{
		MaxRetries:        20,
		InitialDelay:      250 * time.Millisecond,
		MaxDelay:          30 * time.Second,
		BackoffMultiplier: 2.0,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	session, err := client.ConnectTerminalResilient(ctx, target, opts)
	if err != nil {
		fmt.Fprintln(os.Stderr, "connect:", err)
		os.Exit(1)
	}
	session.OnMessage = func(text string) { fmt.Print(text) }
	session.OnReconnect = func(attempt int) { fmt.Printf("\n[reconnected after %d attempts]\n", attempt) }
	session.OnGiveUp = func(err error) { fmt.Printf("\n[retries exhausted: %v]\n", err) }

	_ = session.Send("date\n")
	deadline := time.Now().Add(5 * time.Minute)
	tick := time.NewTicker(30 * time.Second)
	defer tick.Stop()
	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			break
		case <-tick.C:
			_ = session.Send("date\n")
		}
	}
	_ = session.Close()
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

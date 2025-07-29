package main

import (
	"fmt"
	"strings"
	"github.com/shirou/gopsutil/v3/process"
)

func main() {
	fmt.Println("=== Testing CM Process Detection ===")
	
	processes, err := process.Processes()
	if err != nil {
		fmt.Printf("Error getting processes: %v\n", err)
		return
	}
	
	claudeCount := 0
	
	for _, proc := range processes {
		name, err := proc.Name()
		if err != nil {
			continue
		}
		
		if strings.Contains(strings.ToLower(name), "claude") {
			cwd, cwdErr := proc.Cwd()
			if cwdErr != nil {
				cwd = "unknown"
			}
			
			fmt.Printf("Found process: PID=%d, Name='%s', CWD='%s'\n", proc.Pid, name, cwd)
			
			// Check if it's actually a Claude process (not rg, grep, etc.)
			if name == "claude" {
				claudeCount++
				fmt.Printf("  ✅ Valid Claude process\n")
			} else {
				fmt.Printf("  ❌ Not actual Claude (filtered out)\n")
			}
		}
	}
	
	fmt.Printf("\nTotal valid Claude processes: %d\n", claudeCount)
}
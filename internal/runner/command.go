package runner

import (
	"bufio"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/shadow-pulse/shadow-pulse-framework/internal/utils"
)

// LookPath checks if a command exists in the system's PATH.
func LookPath(tool string) bool {
	_, err := exec.LookPath(tool)
	return err == nil
}

// RunCommand executes an external command, shows a progress indicator, and captures output.
func RunCommand(commandString string, useTor bool) error {
	originalCommand := commandString
	// Eyewitness is unstable with proxychains, so we exclude it from Tor routing
	// and rely on its native SOCKS proxy support instead.
	if useTor && !strings.Contains(originalCommand, "eyewitness") {
		if LookPath("proxychains4") {
			commandString = "proxychains4 " + commandString
		} else {
			utils.PrintError("'proxychains4' not found, but --tor flag was used. Cannot route through Tor.")
			return nil // Not a fatal error
		}
	}

	utils.PrintInfo("Executing: " + originalCommand)
	cmd := exec.Command("bash", "-c", commandString)

	stdoutPipe, _ := cmd.StdoutPipe()
	stderrPipe, _ := cmd.StderrPipe()

	if err := cmd.Start(); err != nil {
		utils.PrintError("Failed to start command: " + originalCommand)
		return err
	}

	var outputLines []string
	var mu sync.Mutex
	var wg sync.WaitGroup
	wg.Add(2)

	// Goroutine to read stdout
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stdoutPipe)
		for scanner.Scan() {
			line := scanner.Text()
			mu.Lock()
			outputLines = append(outputLines, line)
			mu.Unlock()
		}
	}()

	// Goroutine to read stderr
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stderrPipe)
		for scanner.Scan() {
			line := scanner.Text()
			mu.Lock()
			outputLines = append(outputLines, "[STDERR] "+line)
			mu.Unlock()
		}
	}()

	// Show progress indicator while the command runs
	done := make(chan struct{})
	go func() {
		for {
			select {
			case <-done:
				return
			default:
				fmt.Print("*")
				time.Sleep(1 * time.Second)
			}
		}
	}()

	err := cmd.Wait()
	close(done)
	fmt.Println() // Newline after progress indicator

	wg.Wait() // Wait for scanner goroutines to finish

	if err != nil {
		utils.PrintError("Command failed: " + originalCommand)
		// Print captured output for debugging
		if len(outputLines) > 0 {
			utils.PrintError("Captured output:\n" + strings.Join(outputLines, "\n"))
		}
		utils.PrintError("Error: " + err.Error())
		return err
	}

	// Optionally, print output on success if needed for some tools,
	// but for now, we'll keep it quiet on success.
	// if len(outputLines) > 0 {
	// 	fmt.Println(strings.Join(outputLines, "\n"))
	// }

	return nil
}

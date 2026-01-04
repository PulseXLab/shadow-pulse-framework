package runner

import (
	"fmt"
	"io"
	"os/exec"
	"time"

	"github.com/shadow-pulse/shadow-pulse-framework/internal/utils"
)

// LookPath checks if a command exists in the system's PATH.
func LookPath(tool string) bool {
	_, err := exec.LookPath(tool)
	return err == nil
}

// RunCommand executes an external command, captures its output, and shows a progress indicator.
func RunCommand(commandString string, useTor bool) error {
	originalCommand := commandString
	
	if useTor {
		if LookPath("proxychains4") {
			commandString = "proxychains4 " + commandString
		} else {
			utils.PrintError("'proxychains4' not found, but --tor flag was used. Cannot route through Tor.")
			return nil // Returning nil to not halt the entire scan for one failed command
		}
	}

	utils.PrintInfo("Executing: " + originalCommand)

	cmd := exec.Command("bash", "-c", commandString)

	    // Start the command
	    stderr, _ := cmd.StderrPipe()
	    if err := cmd.Start(); err != nil {
	        utils.PrintError("Failed to start command: " + originalCommand)
	        return err
	    }
	    
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
	
	    // Wait for the command to finish
	    err := cmd.Wait()
	    close(done)
	    fmt.Println() // Newline after progress indicator
	
	    if err != nil {
	        utils.PrintError("Command failed: " + originalCommand)
	        stderrBytes, _ := io.ReadAll(stderr)
	        if len(stderrBytes) > 0 {
	            utils.PrintError("Stderr: " + string(stderrBytes))
	        }
	        return err
	    }
	return nil
}

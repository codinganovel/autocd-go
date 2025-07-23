# Test Application Examples

This document contains example applications for testing the autocd-go library functionality.

## Simple Directory Navigation Test App

This test application demonstrates the basic usage of `ExitWithDirectory()` and verifies that process replacement works correctly.

### Code

```go
package main

import (
	"fmt"
	"os"
	"github.com/codinganovel/autocd-go"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <directory>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Example: %s /tmp\n", os.Args[0])
		os.Exit(1)
	}

	targetDir := os.Args[1]
	
	fmt.Printf("Current directory: %s\n", getCurrentDir())
	fmt.Printf("Target directory: %s\n", targetDir)
	fmt.Printf("About to call ExitWithDirectory()...\n")
	
	// This should replace the process and never return
	if err := autocd.ExitWithDirectory(targetDir); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	
	// This should NEVER execute if process replacement works
	fmt.Printf("❌ BUG: Process was NOT replaced - this line should never be reached!\n")
	fmt.Printf("Still in directory: %s\n", getCurrentDir())
}

func getCurrentDir() string {
	wd, err := os.Getwd()
	if err != nil {
		return "unknown"
	}
	return wd
}
```

### Usage

1. Create a separate directory for the test app:
   ```bash
   mkdir test_dir
   cd test_dir
   ```

2. Create the test application file (`test_app.go`) with the code above

3. Initialize Go module and add autocd-go dependency:
   ```bash
   go mod init test-app
   go mod edit -replace github.com/codinganovel/autocd-go=../
   go get github.com/codinganovel/autocd-go
   ```

4. Build the application:
   ```bash
   go build -o test_app test_app.go
   ```

5. Test the application:
   ```bash
   ./test_app /tmp
   ```

### Expected Behavior

**Success Case:**
- Displays current directory information
- Calls `ExitWithDirectory()`
- Process is replaced, user gets new shell in target directory
- The "❌ BUG" message never appears
- Shell prompt shows the new directory

**Example Output:**
```
sam@sams-MacBook-Pro test_dir % ./test_app /tmp
Current directory: /Users/sam/Documents/coding/autocd-go/test_dir
Target directory: /tmp
About to call ExitWithDirectory()...
Directory changed to: /tmp
sam@sams-MacBook-Pro /tmp % 
```

**Failure Case:**
- If the library fails, the "❌ BUG" message will appear
- Process continues instead of being replaced
- User remains in original directory

### Testing Different Scenarios

```bash
# Test with valid directory
./test_app /tmp

# Test with invalid directory
./test_app /nonexistent

# Test with current directory
./test_app .

# Test with home directory
./test_app ~
```

## Debug Mode Test

To test debug mode functionality:

```bash
# Without debug mode
./test_app /tmp

# With debug mode enabled
AUTOCD_DEBUG=1 ./test_app /tmp
```

With `AUTOCD_DEBUG=1`, you should see additional debug output like:
```
autocd: platform=1, shell=/bin/zsh (2)
autocd: executing script /tmp/autocd_xxx.sh with shell /bin/zsh
```

## Integration Testing

This test application is useful for:

1. **Verifying library functionality** before integration
2. **Debugging integration issues** by comparing standalone vs integrated behavior  
3. **Testing different environments** and shell configurations
4. **Reproducing reported bugs** in isolated environment

## Notes

- The application must be built in a separate directory to avoid Go package conflicts
- Always test with both valid and invalid directory paths
- The "Directory changed to:" message is normal user feedback, not an error
- Process replacement is the intended behavior - the application should not continue after `ExitWithDirectory()`
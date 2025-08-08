# AutoCD Go - Directory Inheritance for CLI Applications

Finally solve the 50-year Unix limitation: when you navigate directories in a CLI application, inherit that final location in your shell when the app exits.

## The Problem

You're in `/home/user` and run a file manager, editor, or navigation tool. You browse to `/home/user/projects/awesome-project/src/components` and then exit the application. You're back in `/home/user`. You've lost your final location.

This happens with vim, ranger, fzf, and countless other CLI tools. It's been a limitation since the dawn of Unix.

## The Solution

AutoCD Go is a simple library that lets your CLI application "inherit" its final directory to the parent shell. Add one line to your application's exit handler:

```go
autocd.ExitWithDirectory("/path/to/final/directory")
```

When your app exits, the user gets a shell in that directory. Finally.

## Quick Start

### Installation

```bash
go get github.com/codinganovel/autocd-go
```

### Basic Usage

```go
package main

import (
    "github.com/codinganovel/autocd-go"
    "os"
)

func main() {
    // Your application logic here
    // User navigates around, final location is in currentDir
    currentDir := "/path/to/final/directory"
    
    // When ready to exit, inherit the directory
    if err := autocd.ExitWithDirectory(currentDir); err != nil {
        // Fallback to normal exit on any error
        os.Exit(1)
    }
    // This line never executes on success
}
```

### With Command Line Flag

```go
package main

import (
    "flag"
    "os"
    "github.com/codinganovel/autocd-go"
)

func main() {
    var enableAutoCD = flag.Bool("autocd", false, "Change to directory on exit")
    flag.Parse()
    
    app := NewMyApp()
    app.Run()
    
    // Only use autocd if requested and directory changed
    if *enableAutoCD && app.FinalDirectory() != app.StartDirectory() {
        autocd.ExitWithDirectory(app.FinalDirectory())
    }
    
    os.Exit(0)
}
```

## How It Works

The library uses a clever but simple approach:

1. **Creates a POSIX shell script** that changes to your target directory
2. **Replaces the current process** with `/bin/sh` executing this script via `syscall.Exec`
3. **The script changes directory** and then replaces itself with the user's shell
4. **User gets their preferred shell** in the final directory

From the user's perspective, they just run your app and end up where they navigated to.

## Advanced Usage

For more control, use the advanced function:

```go
opts := &autocd.Options{
    SecurityLevel: autocd.SecurityStrict,  // Extra path validation
    DebugMode:     true,                   // Verbose logging
    Shell:         "zsh",                  // Force specific shell
    TempDir:       "/custom/temp",         // Custom temp directory (also cleaned)
}

err := autocd.ExitWithDirectoryAdvanced("/target/path", opts)
```

## Shell Depth Warnings

AutoCD automatically warns when you have many nested shells from navigation:

```
💡 Tip: You have 15 nested shells from navigation.  
For better performance, consider opening a fresh terminal.
```

This feature uses the `SHLVL` environment variable to detect shell nesting depth. Disable warnings if needed:
```go
opts := &autocd.Options{DisableDepthWarnings: true}
```

### Guaranteed Exit

If you want to guarantee your process exits one way or another:

```go
autocd.ExitWithDirectoryOrFallback("/target/path", func() {
    fmt.Println("AutoCD failed, exiting normally")
    os.Exit(0)
})
// Never returns
```

## Platform Support

- **Linux** - bash, zsh, fish, dash, sh
- **macOS** - bash, zsh, fish, dash, sh  
- **BSD** - sh, bash, zsh

The library automatically detects your shell from the `SHELL` environment variable, with automatic fallback to `/bin/sh` if the shell is invalid or missing.

**Note:** AutoCD Go is now focused on Unix-like systems (Linux, macOS, BSD). Windows support has been removed to simplify the architecture and focus on the core Unix use case where directory inheritance is most valuable.

## Security

AutoCD includes built-in security features:

- **Path validation** ensures directories exist and are accessible (execute permission required)
- **Shell injection protection** using single-quote escaping for all paths and shell commands
- **Configurable security levels** from permissive to strict
- **Automatic cleanup** of temporary scripts via periodic removal on subsequent runs

Choose your security level:
- `SecurityNormal` (default) - Path validation, null byte check, directory verification
- `SecurityStrict` - Character whitelist, length limits, comprehensive validation
- `SecurityPermissive` - Minimal validation when you handle security yourself

## Error Handling

The library never crashes your application. On any error, it returns an error and your app can fallback to normal exit:

```go
if err := autocd.ExitWithDirectory("/path"); err != nil {
    fmt.Fprintf(os.Stderr, "Directory inheritance failed: %v\n", err)
    fmt.Fprintf(os.Stderr, "Final directory was: %s\n", "/path")
    os.Exit(0)  // Normal exit
}
```

## Real-World Examples

### File Manager
```go
// When user presses 'q' to quit
if fm.autoCDEnabled && fm.currentDir != fm.startDir {
    autocd.ExitWithDirectory(fm.currentDir)
}
os.Exit(0)
```

### Directory Picker
```go
// After user selects a directory
selectedDir := showDirectoryPicker()
autocd.ExitWithDirectory(selectedDir)
```

### Interactive Shell/REPL
```go
// When user types 'exit'
currentWorkingDir, _ := os.Getwd()
autocd.ExitWithDirectory(currentWorkingDir)
```

## Environment Variables

- `AUTOCD_DEBUG=1` - Enable debug output
- `SHELL` - Override shell detection

## Dependencies

None. AutoCD uses only the Go standard library.

## Technical Details

For comprehensive technical documentation, implementation details, and advanced integration patterns, see [readme-developers.md](readme-developers.md).

## Contributing

This library is designed to be simple and robust. The core concept is stable, but we welcome:

- Bug reports and fixes
- Additional shell support
- Platform compatibility improvements
- Documentation improvements

## License

under ☕️, check out [the-coffee-license](https://github.com/codinganovel/The-Coffee-License)

I've included both licenses with the repo, do what you know is right. The licensing works by assuming you're operating under good faith.

## Finally

After 50 years, CLI applications can finally tell the shell where they've been. Your users will love you for it.

---

*"It's about time." - Every terminal user ever*
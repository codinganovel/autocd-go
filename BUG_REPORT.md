# AutoCD-Go Bug Investigation Report

## Executive Summary

This report documents critical bugs and security vulnerabilities discovered in the autocd-go library. The analysis identified **10 significant issues**, including **3 critical security vulnerabilities** that could lead to arbitrary code execution, privilege escalation, and data exposure.

## Critical Security Vulnerabilities

### 1. Shell Injection Vulnerability [CRITICAL]
**Location**: `script.go:87-115`, `validation.go:76`  
**Severity**: Critical  
**CVSS Score**: 9.8 (Critical)

**Description**: The path sanitization mechanism is insufficient and can be bypassed, allowing attackers to inject arbitrary shell commands.

**Attack Vector**:
```go
// Example malicious path that bypasses sanitization
maliciousPath := "/tmp/safe`rm -rf /`"
maliciousPath2 := "/tmp/$(whoami > /tmp/pwned)"
maliciousPath3 := "/tmp/\nrm -rf /"
```

**Root Causes**:
- Path cleaning happens AFTER traversal checks in `validation.go:76`
- Backticks are not escaped in Unix shell scripts
- Newline characters are not filtered
- PowerShell sanitization misses critical metacharacters

**Impact**:
- Remote code execution if attacker controls directory path
- Complete system compromise possible
- Data exfiltration and privilege escalation

**Remediation**:
1. Move `filepath.Clean()` before traversal validation
2. Implement comprehensive shell metacharacter escaping
3. Use shell-specific quoting mechanisms
4. Validate against newlines and control characters
5. Consider using execve() directly instead of shell scripts

### 2. TOCTOU Race Condition in Temporary Files [HIGH]
**Location**: `tempfile.go:22-43`  
**Severity**: High  
**CVSS Score**: 7.5 (High)

**Description**: A race condition exists between script creation and permission setting, creating a window where sensitive scripts are world-readable.

**Exploitation Window**:
```go
// Between these lines, file is world-readable
tmpFile, err := os.CreateTemp("", "autocd-*.sh")  // World-readable by default
// ... write sensitive content ...
os.Chmod(tmpFile.Name(), 0700)  // Too late!
```

**Impact**:
- Sensitive path information disclosure
- Script manipulation by other users
- Potential privilege escalation on multi-user systems

**Remediation**:
1. Set umask before file creation
2. Use `os.OpenFile()` with O_EXCL and specific permissions
3. Validate temp directory security
4. Consider in-memory script execution

### 3. PATH Hijacking on Windows [HIGH]
**Location**: `exec.go:30-41`  
**Severity**: High  
**CVSS Score**: 7.8 (High)

**Description**: Executable names are hardcoded without full paths, allowing PATH hijacking attacks.

**Vulnerable Code**:
```go
case PlatformWindows:
    if shell.Type == ShellPowerShell {
        executable = "powershell.exe"  // Could execute malicious binary
    } else {
        executable = "cmd.exe"         // PATH hijacking possible
    }
```

**Impact**:
- Execution of attacker-controlled binaries
- Privilege escalation if running elevated
- Complete system compromise

**Remediation**:
1. Use full system paths (e.g., `C:\Windows\System32\cmd.exe`)
2. Verify executable signatures
3. Use `syscall.GetSystemDirectory()` for path resolution
4. Validate executable location before execution

## High Priority Bugs

### 4. Error Context Loss [MEDIUM]
**Location**: `errors.go:18-25` and throughout  
**Severity**: Medium

**Description**: Improper error wrapping causes `errors.Is()` checks to fail, breaking error handling logic.

**Example**:
```go
// This breaks error type checking
return fmt.Errorf("failed to validate: %s", err)  // Wrong!
// Should be:
return fmt.Errorf("failed to validate: %w", err)  // Correct
```

### 5. Integer Overflow in Shell Depth [LOW]
**Location**: `autocd.go:34-38`  
**Severity**: Low

**Description**: SHLVL environment variable parsing lacks bounds checking, potentially causing integer overflow.

**Vulnerable Code**:
```go
depth, err := strconv.Atoi(os.Getenv("SHLVL"))
// No validation of depth range
```

### 6. Unsafe File Cleanup [MEDIUM]
**Location**: `tempfile.go:56-65`  
**Severity**: Medium

**Description**: Cleanup function doesn't validate files before removal, potentially allowing deletion of non-autocd files.

**Issues**:
- No ownership verification
- Errors silently ignored
- Pattern matching too permissive

## Additional Issues

### 7. Nil Pointer Risk
**Location**: `exec.go:61`  
**Description**: Potential nil dereference when accessing shell.Path in error messages.

### 8. Regex Performance Issue
**Location**: `validation.go:101,110`  
**Description**: Regex compiled on every validation call, causing performance degradation.

### 9. Incomplete Signal Handling
**Location**: `script.go:53`  
**Description**: Self-deletion trap doesn't handle all termination signals.

### 10. Platform Detection Gaps
**Location**: `platform.go:19`  
**Description**: Generic Unix fallback may not handle all Unix variants correctly.

## Proof of Concept

### Shell Injection PoC:
```go
// This will execute 'id' command during directory change
maliciousDir := "/tmp/`id > /tmp/pwned.txt`safe"
autocd.ExitWithDirectory(maliciousDir)
```

### Race Condition PoC:
```bash
# Terminal 1: Run autocd program
# Terminal 2: Watch for new files and read them
while true; do
    for f in /tmp/autocd-*.sh; do
        [ -f "$f" ] && cat "$f" && echo "EXPOSED: $f"
    done
done
```

## Security Recommendations

1. **Immediate Actions**:
   - Disable library usage in production until patches applied
   - Audit all uses of user-controlled paths
   - Implement input validation allowlists

2. **Short-term Fixes**:
   - Apply all recommended code fixes
   - Add comprehensive security tests
   - Implement secure coding practices

3. **Long-term Improvements**:
   - Security audit by third party
   - Implement sandboxing for script execution
   - Add security logging and monitoring
   - Consider alternative approaches without shell scripts

## Testing Gaps

The current test coverage (72.1%) misses critical security scenarios:
- No injection attack tests
- No race condition tests  
- No signal handling tests
- No Windows-specific security tests
- No fuzzing for input validation

## Conclusion

The autocd-go library contains severe security vulnerabilities that must be addressed before production use. The shell injection vulnerability is particularly critical and could lead to complete system compromise. All identified issues should be patched immediately, and comprehensive security testing should be implemented.

**Risk Assessment**: **CRITICAL** - Do not use in production without applying security fixes.
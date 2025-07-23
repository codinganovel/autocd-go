# Error Handling Improvement Idea

## Problem: Brittle Error Logic

Currently, the library's internal error handling relies on parsing the text of error messages to determine the specific type of error.

For example, the `validateTargetPath` function returns a generic error with a message like `"path is not a directory"`. The `newPathValidationError` function then inspects this error by checking if its string content contains the substring `"not a directory"`.

This creates a fragile ("brittle") connection between two separate parts of the library. If a developer decides to improve or change the error message text in `validateTargetPath`, the string-checking logic in `newPathValidationError` will fail, introducing a bug. The program's logic should not depend on human-readable strings.

## Proposed Implementation Plan

To make the error handling more robust and align with modern Go best practices, we should refactor it to use specific error values instead of string matching.

1.  **Define Exported Error Variables:** In `errors.go` or `validation.go`, define and export specific variables for each validation failure.
    ```go
    var ErrPathNotFound = errors.New("path does not exist")
    var ErrPathNotDirectory = errors.New("path is not a directory")
    var ErrPathNotAccessible = errors.New("path is not accessible")
    // etc.
    ```

2.  **Return Specific Errors:** Refactor the `validateTargetPath` function in `validation.go` to return these newly defined error variables directly, instead of creating a new error with `fmt.Errorf`.
    ```go
    // Before
    // return "", fmt.Errorf("path is not a directory: %s", absPath)

    // After
    return "", ErrPathNotDirectory
    ```

3.  **Use `errors.Is()` for Checks:** Refactor the `newPathValidationError` function in `errors.go` to use `errors.Is()` to check for the specific error variables. This removes the dependency on string content.
    ```go
    // Before
    // if strings.Contains(cause.Error(), "not a directory") { ... }

    // After
    if errors.Is(cause, ErrPathNotDirectory) {
        errType = ErrorPathNotDirectory
    }
    ```

This change will decouple the error's identity from its description, allowing the error messages to be changed freely without breaking the application's logic.

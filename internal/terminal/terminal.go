// Package terminal provides cross-platform terminal raw mode utilities
package terminal

import (
    "golang.org/x/term"
)

// MakeRaw puts the terminal into raw mode for single-character input
// This allows reading input without waiting for Enter key
func MakeRaw(fd int) (*term.State, error) {
    return term.MakeRaw(fd) // Automatically handles platform differences
}

// Restore restores the terminal to its original state
func Restore(fd int, oldState *term.State) error {
    return term.Restore(fd, oldState)
}

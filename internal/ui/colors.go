package ui

import (
	"fmt"
	"io"

	"github.com/fatih/color"
)

var (
	// Success - green
	Success = color.New(color.FgGreen).SprintFunc()
	// Error - red
	Error = color.New(color.FgRed).SprintFunc()
	// Warning - yellow
	Warning = color.New(color.FgYellow).SprintFunc()
	// Info - cyan
	Info = color.New(color.FgCyan).SprintFunc()
	// Highlight - bold
	Highlight = color.New(color.Bold).SprintFunc()
	// Dim - faint
	Dim = color.New(color.Faint).SprintFunc()
)

// Printer provides colored output functions
type Printer struct {
	writer io.Writer
	noColor bool
}

// NewPrinter creates a new printer
func NewPrinter(w io.Writer, noColor bool) *Printer {
	color.NoColor = noColor
	return &Printer{
		writer: w,
		noColor: noColor,
	}
}

// Success prints success message in green
func (p *Printer) Success(format string, args ...interface{}) {
	fmt.Fprintf(p.writer, Success(format)+"\n", args...)
}

// Error prints error message in red
func (p *Printer) Error(format string, args ...interface{}) {
	fmt.Fprintf(p.writer, Error(format)+"\n", args...)
}

// Warning prints warning message in yellow
func (p *Printer) Warning(format string, args ...interface{}) {
	fmt.Fprintf(p.writer, Warning(format)+"\n", args...)
}

// Info prints info message in cyan
func (p *Printer) Info(format string, args ...interface{}) {
	fmt.Fprintf(p.writer, Info(format)+"\n", args...)
}

// Println prints normal message
func (p *Printer) Println(format string, args ...interface{}) {
	fmt.Fprintf(p.writer, format+"\n", args...)
}

package banner

import (
	"fmt"
	"os"
)

// Field represents a key-value pair in the banner.
type Field struct {
	Key   string
	Value string
}

// Print writes a standard ASCII banner to stdout.
// It is used to present a human-readable startup summary in the terminal
// and avoids polluting the structured log stream.
func Print(name, version string, fields []Field) {
	_, _ = fmt.Fprintf(os.Stdout, "\n")
	_, _ = fmt.Fprintf(os.Stdout, "  ┌──────────────────────────────────────────────────┐\n")
	_, _ = fmt.Fprintf(os.Stdout, "  │  %-48s│\n", name+" "+version)
	_, _ = fmt.Fprintf(os.Stdout, "  ├──────────────────────────────────────────────────┤\n")
	for _, f := range fields {
		_, _ = fmt.Fprintf(os.Stdout, "  │  %-10s %-37s│\n", f.Key, f.Value)
	}
	_, _ = fmt.Fprintf(os.Stdout, "  └──────────────────────────────────────────────────┘\n")
	_, _ = fmt.Fprintf(os.Stdout, "\n")
}

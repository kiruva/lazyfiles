package fileops

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// format identifies an archive type detected from a filename.
type format int

const (
	fmtUnknown format = iota
	fmtTar
	fmtTarGz
	fmtTarBz2
	fmtTarXz
	fmtTarZst
	fmtZip
	fmt7z
	fmtRar
)

// suffixes are checked longest-first so ".tar.gz" wins over ".gz".
var suffixes = []struct {
	ext string
	f   format
}{
	{".tar.gz", fmtTarGz}, {".tgz", fmtTarGz},
	{".tar.bz2", fmtTarBz2}, {".tbz2", fmtTarBz2}, {".tbz", fmtTarBz2},
	{".tar.xz", fmtTarXz}, {".txz", fmtTarXz},
	{".tar.zst", fmtTarZst}, {".tzst", fmtTarZst},
	{".tar", fmtTar},
	{".zip", fmtZip},
	{".7z", fmt7z},
	{".rar", fmtRar},
}

func detectFormat(name string) format {
	lower := strings.ToLower(name)
	for _, s := range suffixes {
		if strings.HasSuffix(lower, s.ext) {
			return s.f
		}
	}
	return fmtUnknown
}

// IsArchive reports whether name looks like a supported archive.
func IsArchive(name string) bool { return detectFormat(name) != fmtUnknown }

// tarComp maps a tar format to its compression flag (empty for plain tar).
func tarComp(f format) string {
	switch f {
	case fmtTarGz:
		return "-z"
	case fmtTarBz2:
		return "-j"
	case fmtTarXz:
		return "-J"
	case fmtTarZst:
		return "--zstd"
	default:
		return ""
	}
}

func isTar(f format) bool {
	return f == fmtTar || f == fmtTarGz || f == fmtTarBz2 || f == fmtTarXz || f == fmtTarZst
}

// pack creates job.Out from job.Srcs. MVP always writes a gzip-compressed tar.
// All sources come from the same pane, so they share a parent directory.
func pack(job Job, r *reporter) error {
	if len(job.Srcs) == 0 {
		return fmt.Errorf("nothing to pack")
	}
	parent := filepath.Dir(job.Srcs[0])
	args := []string{"-c", "-z", "-v", "-f", job.Out, "-C", parent}
	for _, s := range job.Srcs {
		args = append(args, filepath.Base(s))
	}
	return runStreaming("tar", args, r, parseTarLine)
}

// extractAll extracts each archive in job.Srcs into job.Dest.
func extractAll(job Job, r *reporter) error {
	if err := os.MkdirAll(job.Dest, 0o755); err != nil {
		return err
	}
	for _, arc := range job.Srcs {
		if err := extractOne(arc, job.Dest, r); err != nil {
			return err
		}
	}
	return nil
}

func extractOne(arc, dest string, r *reporter) error {
	f := detectFormat(arc)
	bin, args, parse := extractCommand(f, arc, dest)
	if bin == "" {
		return fmt.Errorf("unsupported archive: %s", filepath.Base(arc))
	}
	return runStreaming(bin, args, r, parse)
}

// extractCommand returns the binary, args, and line parser to extract arc into dest.
func extractCommand(f format, arc, dest string) (bin string, args []string, parse func(string) string) {
	switch {
	case isTar(f):
		a := []string{"-x"}
		if c := tarComp(f); c != "" {
			a = append(a, c)
		}
		a = append(a, "-v", "-f", arc, "-C", dest)
		return "tar", a, parseTarLine
	case f == fmtZip:
		return "unzip", []string{"-o", arc, "-d", dest}, parseUnzipLine
	case f == fmt7z:
		return "7z", []string{"x", arc, "-o" + dest, "-y"}, parseNone
	case f == fmtRar:
		return "unrar", []string{"x", "-y", arc, dest + string(os.PathSeparator)}, parseNone
	default:
		return "", nil, nil
	}
}

// countArchiveEntries returns the number of members in an archive, or 0 if it
// cannot be determined cheaply (progress falls back to indeterminate).
func countArchiveEntries(arc string) int {
	f := detectFormat(arc)
	switch {
	case isTar(f):
		a := []string{"-t"}
		if c := tarComp(f); c != "" {
			a = append(a, c)
		}
		a = append(a, "-f", arc)
		return countLines(exec.Command("tar", a...))
	case f == fmtZip:
		return countLines(exec.Command("zipinfo", "-1", arc))
	default:
		return 0
	}
}

func countLines(cmd *exec.Cmd) int {
	out, err := cmd.Output()
	if err != nil {
		return 0
	}
	return bytes.Count(bytes.TrimRight(out, "\n"), []byte{'\n'}) + boolToInt(len(bytes.TrimSpace(out)) > 0)
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// runStreaming runs a command, merging stdout+stderr into one stream, and calls
// r.step for each output line the parser deems a processed file. Using a single
// *os.File pipe for both streams keeps their lines in order and applies natural
// backpressure: if the UI is slow to drain progress, the child process blocks.
func runStreaming(bin string, args []string, r *reporter, parse func(string) string) error {
	cmd := exec.Command(bin, args...)
	pr, pw, err := os.Pipe()
	if err != nil {
		return err
	}
	cmd.Stdout = pw
	cmd.Stderr = pw

	if err := cmd.Start(); err != nil {
		pw.Close()
		pr.Close()
		return err
	}
	pw.Close() // parent drops the write end; reader sees EOF once the child exits

	sc := bufio.NewScanner(pr)
	sc.Buffer(make([]byte, 64*1024), 1024*1024)
	for sc.Scan() {
		if name := parse(sc.Text()); name != "" {
			r.step(name)
		}
	}
	pr.Close()

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("%s failed: %w", bin, err)
	}
	return nil
}

// Line parsers ----------------------------------------------------------------

// parseTarLine: `tar -v` prints one member path per line.
func parseTarLine(s string) string { return strings.TrimSpace(s) }

// parseUnzipLine: `unzip` prints "  inflating: path", "extracting: path", etc.
func parseUnzipLine(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.Index(s, ": "); i >= 0 {
		return strings.TrimSpace(s[i+2:])
	}
	return ""
}

// parseNone drives an indeterminate bar (tool output not worth parsing).
func parseNone(string) string { return "" }

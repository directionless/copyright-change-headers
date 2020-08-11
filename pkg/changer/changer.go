package changer

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/tabwriter"
)

type langType string

const (
	CStyle       langType = "c"
	PyStyle               = "python"
	ShStyle               = "shell"
	explicitFile          = "explicit"
	unknownStyle          = "unknown"
	ignoreStyle           = "ignored"
)

type replacer interface {
	Replace(s string) string
	WriteString(w io.Writer, s string) (n int, err error)
}

type changer struct {
	baseLicense    []string
	count          map[langType]int
	ignoredFiles   []string
	license        map[langType]string
	oldLicenses    map[langType][]string
	regexpCleaners map[*regexp.Regexp][]byte
	replacers      map[langType]replacer
}

type Opts func(*changer)

func WithOldLicense(style langType, lic []byte) Opts {
	return func(c *changer) {
		c.oldLicenses[style] = append(c.oldLicenses[style], string(lic))
	}
}

func WithRegexCleaner(re *regexp.Regexp, repl []byte) Opts {
	return func(c *changer) {
		c.regexpCleaners[re] = repl
	}

}

func WithIgnoredFile(f string) Opts {
	return func(c *changer) {
		c.ignoredFiles = append(c.ignoredFiles, f)
	}
}

func New(baseLicense []string, opts ...Opts) *changer {
	c := &changer{
		baseLicense:    baseLicense,
		replacers:      make(map[langType]replacer),
		count:          make(map[langType]int),
		oldLicenses:    make(map[langType][]string),
		license:        make(map[langType]string),
		regexpCleaners: make(map[*regexp.Regexp][]byte),
	}
	c.license[CStyle] = c.formatLicense(`/**`, ` * `, ` */`, 1)
	c.license[PyStyle] = c.formatLicense(``, `# `, ``, 1)
	c.license[ShStyle] = c.formatLicense(``, `# `, ``, 1)

	for _, opt := range opts {
		opt(c)
	}

	for style, oldLicenses := range c.oldLicenses {
		oldnew := []string{}
		for _, lic := range oldLicenses {
			// Add with different numbers of newlines, so that we collapse them
			oldnew = append(oldnew, lic+"\n\n", c.license[style])
			oldnew = append(oldnew, lic+"\n", c.license[style])
			oldnew = append(oldnew, lic, c.license[style])
		}
		c.replacers[style] = strings.NewReplacer(oldnew...)
	}

	return c
}

func (c *changer) Status(w io.Writer) {
	tw := tabwriter.NewWriter(w, 0, 2, 2, ' ', 0)

	fmt.Fprintf(tw, "Examined files by type:\n")
	total := 0
	for style, count := range c.count {
		fmt.Fprintf(tw, "%s\t%d\n", style, count)
		total = total + count
	}
	fmt.Fprintf(tw, "\t\n")
	fmt.Fprintf(tw, "%s\t%d\n", "total", total)
	tw.Flush()
}

func (c *changer) Walk(dir string) error {
	return filepath.Walk(dir, c.WalkFn)
}

func (c *changer) WalkFn(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}

	style := c.styleClassifier(path)
	c.count[style] = c.count[style] + 1

	//fmt.Println(path, " ", style)

	if style == unknownStyle || style == ignoreStyle {
		return nil
	}

	r, ok := c.replacers[style]
	if !ok {
		return fmt.Errorf("No replacer for %s", style)
	}

	input, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading %s: %w", input, err)
	}

	for re, repl := range c.regexpCleaners {
		input = re.ReplaceAll(input, repl)
	}

	output, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("opening %s for writing: %w", input, err)
	}
	defer output.Close()

	if _, err := r.WriteString(output, string(input)); err != nil {
		return fmt.Errorf("replacing style %s: %w", style, err)
	}

	if err := output.Close(); err != nil {
		return fmt.Errorf("closing %s: %w", path, err)
	}

	return nil
}

func (c *changer) styleClassifier(path string) langType {

	for _, ig := range c.ignoredFiles {
		if strings.HasSuffix(path, ig) {
			return ignoreStyle
		}
	}

	switch filepath.Base(path) {
	case "CMakeLists.txt":
		return ShStyle
	}

	switch filepath.Ext(strings.TrimSuffix(path, ".in")) {
	case "", ".table", ".md", ".json", ".xml", ".debian":
		return ignoreStyle
	case ".c", ".cpp", ".h", ".hpp", ".mm":
		return CStyle
	case ".py":
		return PyStyle
	case ".sh", ".ps1", ".cmake":
		return ShStyle
	}

	return unknownStyle
}

func (c *changer) formatLicense(header, indent, footer string, blankCount int) string {
	var b strings.Builder
	if header != "" {
		fmt.Fprintf(&b, "%s\n", header)
	}
	for _, line := range c.baseLicense {
		fmt.Fprintf(&b, "%s\n", strings.TrimRight(indent+line, " "))

	}
	if footer != "" {
		fmt.Fprintf(&b, "%s\n", footer)
	}

	for i := 0; i < blankCount; i++ {
		fmt.Fprintf(&b, "\n")
	}

	return b.String()
}

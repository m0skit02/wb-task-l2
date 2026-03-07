package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"
)

type options struct {
	after  int
	before int
	count  bool
	ignore bool
	invert bool
	fixed  bool
	number bool
}

func parseFlags() (options, string, string) {
	var opt options
	var context int

	flag.IntVar(&opt.after, "A", 0, "print N lines after match")
	flag.IntVar(&opt.before, "B", 0, "print N lines before match")
	flag.IntVar(&context, "C", 0, "print N lines of context")
	flag.BoolVar(&opt.count, "c", false, "print count only")
	flag.BoolVar(&opt.ignore, "i", false, "ignore case")
	flag.BoolVar(&opt.invert, "v", false, "invert match")
	flag.BoolVar(&opt.fixed, "F", false, "fixed string")
	flag.BoolVar(&opt.number, "n", false, "print line numbers")

	flag.Parse()

	if context > 0 {
		opt.after = context
		opt.before = context
	}

	args := flag.Args()

	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "pattern required")
		os.Exit(1)
	}

	pattern := args[0]

	var file string
	if len(args) > 1 {
		file = args[1]
	}

	return opt, pattern, file
}

func readLines(file string) ([]string, error) {
	var scanner *bufio.Scanner

	if file == "" {
		scanner = bufio.NewScanner(os.Stdin)
	} else {
		f, err := os.Open(file)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		scanner = bufio.NewScanner(f)
	}

	var lines []string

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return lines, scanner.Err()
}

func buildMatcher(pattern string, opt options) (func(string) bool, error) {
	if opt.fixed {
		if opt.ignore {
			pattern = strings.ToLower(pattern)
		}

		return func(line string) bool {
			if opt.ignore {
				line = strings.ToLower(line)
			}
			return strings.Contains(line, pattern)
		}, nil
	}

	if opt.ignore {
		pattern = "(?i)" + pattern
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}

	return func(line string) bool {
		return re.MatchString(line)
	}, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func main() {
	opt, pattern, file := parseFlags()

	lines, err := readLines(file)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	match, err := buildMatcher(pattern, opt)
	if err != nil {
		fmt.Fprintln(os.Stderr, "invalid regexp:", err)
		os.Exit(1)
	}

	matched := make([]bool, len(lines))
	matchCount := 0

	for i, line := range lines {
		ok := match(line)

		if opt.invert {
			ok = !ok
		}

		if ok {
			matched[i] = true
			matchCount++
		}
	}

	if opt.count {
		fmt.Println(matchCount)
		return
	}

	printed := make(map[int]bool)

	for i := range lines {
		if !matched[i] {
			continue
		}

		start := max(0, i-opt.before)
		end := min(len(lines)-1, i+opt.after)

		for j := start; j <= end; j++ {
			if printed[j] {
				continue
			}

			printed[j] = true

			if opt.number {
				fmt.Printf("%d:%s\n", j+1, lines[j])
			} else {
				fmt.Println(lines[j])
			}
		}
	}
}

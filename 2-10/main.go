package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

type options struct {
	column  int
	numeric bool
	reverse bool
	unique  bool
	month   bool
	trim    bool
	check   bool
	human   bool
}

func main() {
	opts := parseFlags()

	lines, err := readInput(flag.Args())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if opts.trim {
		for i := range lines {
			lines[i] = strings.TrimRight(lines[i], " \t")
		}
	}

	lessFunc := buildComparator(lines, opts)

	if opts.check {
		if !sort.SliceIsSorted(lines, lessFunc) {
			fmt.Println("Data is not sorted")
			os.Exit(1)
		}
		return
	}

	sort.SliceStable(lines, lessFunc)

	if opts.unique {
		lines = unique(lines)
	}

	for _, line := range lines {
		fmt.Println(line)
	}
}

func parseFlags() options {
	var opts options

	flag.IntVar(&opts.column, "k", 0, "sort by column")
	flag.BoolVar(&opts.numeric, "n", false, "numeric sort")
	flag.BoolVar(&opts.reverse, "r", false, "reverse sort")
	flag.BoolVar(&opts.unique, "u", false, "unique lines")
	flag.BoolVar(&opts.month, "M", false, "month sort")
	flag.BoolVar(&opts.trim, "b", false, "trim trailing blanks")
	flag.BoolVar(&opts.check, "c", false, "check sorted")
	flag.BoolVar(&opts.human, "h", false, "human numeric sort")

	flag.Parse()

	return opts
}

func readInput(files []string) ([]string, error) {
	var scanner *bufio.Scanner

	if len(files) > 0 {
		file, err := os.Open(files[0])
		if err != nil {
			return nil, err
		}
		defer file.Close()
		scanner = bufio.NewScanner(file)
	} else {
		scanner = bufio.NewScanner(os.Stdin)
	}

	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return lines, scanner.Err()
}

func buildComparator(lines []string, opts options) func(i, j int) bool {
	return func(i, j int) bool {
		a := extractKey(lines[i], opts)
		b := extractKey(lines[j], opts)

		var result bool

		switch {
		case opts.month:
			result = monthValue(a) < monthValue(b)
		case opts.human:
			result = humanValue(a) < humanValue(b)
		case opts.numeric:
			af, _ := strconv.ParseFloat(a, 64)
			bf, _ := strconv.ParseFloat(b, 64)
			result = af < bf
		default:
			result = a < b
		}

		if opts.reverse {
			return !result
		}

		return result
	}
}

func extractKey(line string, opts options) string {
	if opts.column <= 0 {
		return line
	}

	fields := strings.Split(line, "\t")
	if opts.column-1 < len(fields) {
		return fields[opts.column-1]
	}
	return ""
}

func unique(lines []string) []string {
	if len(lines) == 0 {
		return lines
	}

	result := []string{lines[0]}

	for i := 1; i < len(lines); i++ {
		if lines[i] != lines[i-1] {
			result = append(result, lines[i])
		}
	}
	return result
}

var months = map[string]int{
	"Jan": 1, "Feb": 2, "Mar": 3,
	"Apr": 4, "May": 5, "Jun": 6,
	"Jul": 7, "Aug": 8, "Sep": 9,
	"Oct": 10, "Nov": 11, "Dec": 12,
}

func monthValue(s string) int {
	if val, ok := months[s]; ok {
		return val
	}
	return 0
}

func humanValue(s string) float64 {
	multiplier := 1.0
	last := s[len(s)-1]

	switch last {
	case 'K', 'k':
		multiplier = 1 << 10
		s = s[:len(s)-1]
	case 'M', 'm':
		multiplier = 1 << 20
		s = s[:len(s)-1]
	case 'G', 'g':
		multiplier = 1 << 30
		s = s[:len(s)-1]
	}

	value, _ := strconv.ParseFloat(s, 64)
	return value * multiplier
}

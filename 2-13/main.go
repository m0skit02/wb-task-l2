package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func parseFields(s string) map[int]struct{} {
	fields := make(map[int]struct{})

	parts := strings.Split(s, ",")

	for _, p := range parts {
		if strings.Contains(p, "-") {
			r := strings.Split(p, "-")

			start, _ := strconv.Atoi(r[0])
			end, _ := strconv.Atoi(r[1])

			for i := start; i <= end; i++ {
				fields[i] = struct{}{}
			}

		} else {
			val, _ := strconv.Atoi(p)
			fields[val] = struct{}{}
		}
	}

	return fields
}

func main() {

	fieldsFlag := flag.String("f", "", "fields")
	delimiter := flag.String("d", "\t", "delimiter")
	separated := flag.Bool("s", false, "only separated")

	flag.Parse()

	if *fieldsFlag == "" {
		fmt.Println("flag -f is required")
		os.Exit(1)
	}

	fields := parseFields(*fieldsFlag)

	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {

		line := scanner.Text()

		if *separated && !strings.Contains(line, *delimiter) {
			continue
		}

		cols := strings.Split(line, *delimiter)

		var result []string

		for i, col := range cols {

			if _, ok := fields[i+1]; ok {
				result = append(result, col)
			}
		}

		if len(result) > 0 {
			fmt.Println(strings.Join(result, *delimiter))
		}
	}
}

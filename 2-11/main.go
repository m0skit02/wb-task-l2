package main

import (
	"fmt"
	"sort"
	"strings"
)

func FindAnagrams(words []string) map[string][]string {
	groups := make(map[string][]string)
	firstWord := make(map[string]string)

	for _, word := range words {
		lower := strings.ToLower(word)

		runes := []rune(lower)
		sort.Slice(runes, func(i, j int) bool {
			return runes[i] < runes[j]
		})

		key := string(runes)

		if _, ok := firstWord[key]; !ok {
			firstWord[key] = lower
		}

		groups[key] = append(groups[key], lower)
	}

	result := make(map[string][]string)

	for key, group := range groups {
		if len(group) < 2 {
			continue
		}

		sort.Strings(group)
		result[firstWord[key]] = group
	}

	return result
}

func main() {
	input := []string{
		"пятак", "пятка", "тяпка",
		"листок", "слиток", "столик",
		"стол",
	}

	result := FindAnagrams(input)

	for k, v := range result {
		fmt.Println(k, "->", v)
	}
}

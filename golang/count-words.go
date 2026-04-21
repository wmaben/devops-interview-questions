package main

import (
	"fmt"
)

func countAlphabets(sentence string) int {
	count := 0
	for _, word := range sentence {
		if word == ' ' || word == '\t' || word == '\n' {
			continue
		} else {
			count++
		}
	}

	return count
}

func countWords(sentence string) int {
	count := 0
	var isWord = true
	for _, word := range sentence {
		if word == ' ' || word == '\t' || word == '\n' {
			isWord = true
		} else {
			if isWord {
				count++
				isWord = false
			}
		}
	}

	return count
}

func main() {
	sentence := "Hello r temp"
	fmt.Printf("Number of alphabets in string is %d\n", countAlphabets(sentence))
	fmt.Printf("Number of words in string is %d\n", countWords(sentence))
}

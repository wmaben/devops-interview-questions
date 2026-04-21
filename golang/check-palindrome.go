// **Problem:** Given a string, determine if it is a palindrome, considering only alphanumeric characters and ignoring cases.

package main

import (
	"fmt"
	"unicode"
)

func isPalindrome(s string) bool {
	runes := []rune(s)
	for _, char := range s {
		if unicode.IsLetter(char) || unicode.IsDigit(char) {
			runes = append(runes, unicode.ToLower(char))
		}
	}

	i, j := 0, len(runes)-1
	for i < j {
		runes[i], runes[j] = runes[j], runes[i]
		i++
		j--
	}
	if string(runes) != s {
		return false
	}
	return true
}

func main() {
	fmt.Println(isPalindrome("12234"))
}

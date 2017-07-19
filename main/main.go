package main

import (
	"bytes"
	"fmt"
	"github.com/pkg/profile"
	"io"
	"levenshteinSearch/levenshtein"
	"os"
	_ "runtime/pprof"
	"strconv"
	"strings"
	"time"
)

const (
	vacabularyName = "vocabulary.txt"
	buffer         = 100
	minShift       = 1
)

var (
	maxLength  = 0
	vocabulary map[int][]string
	cache      map[string]int
)

func main() {
	defer profile.Start(profile.CPUProfile).Stop()
	start := time.Now()
	filename := os.Args[1]
	buf := bytes.NewBuffer(nil)
	file, err := os.Open(filename)
	defer file.Close()
	if err != nil {
		panic(fmt.Sprintf("File with name '%s' not found", filename))
	}
	io.Copy(buf, file)

	str := strings.ToUpper(string(buf.Bytes()))
	if len(str) == 0 {
		panic(fmt.Sprintf("Empty file!"))
	}
	cache = make(map[string]int)
	before := len(str)
	newLen := 0
	for before != newLen {
		before = len(str)
		str = strings.Replace(str, "\n", "", -1)
		str = strings.Replace(str, "  ", " ", -1)
		newLen = len(str)
	}
	words := strings.Split(str, " ")
	vocabulary, err = prepareVocabulary()
	if err != nil {
		panic(err)
	}

	in := make(chan string, buffer)
	out := make(chan int, buffer)
	for i := 0; i < len(words); i++ {
		go getDistance(in, out)
	}
	for _, word := range words {
		in <- word
	}
	close(in)
	distance := 0
	for range words {
		result := <-out
		distance += result
	}
	close(out)
	fmt.Println("Result: " + strconv.Itoa(distance))
	elapsed := time.Since(start)
	fmt.Println("Calculation took ", elapsed)
}

// prepareVocabulary return vocabulary as slice
func prepareVocabulary() (map[int][]string, error) {
	buf := bytes.NewBuffer(nil)
	file, err := os.Open(vacabularyName)
	defer file.Close()
	if err != nil {
		return nil, err
	}
	io.Copy(buf, file)
	str := strings.TrimSpace(string(buf.Bytes()))
	words := strings.Split(str, "\n")
	wordsSort := make(map[int][]string)
	for _, word := range words {
		max := len(word)
		wordsSort[max] = append(wordsSort[len(word)], word)
		if max > maxLength {
			maxLength = max
		}
	}
	return wordsSort, nil
}

func getDistance(in <-chan string, out chan<- int) {
	word := <-in
	if min, ok := cache[word]; ok {
		out <- min
		return
	}
	length := len(word)
	min := 999
	shiftLeft := length
	shiftRight := length
loop:
	for {
		if shiftLeft < minShift || shiftRight > maxLength {
			break loop
		}
		diff := shiftLeft - length
		if min == diff {
			break loop
		}
		if shiftLeft == length {
			lengthVoc := length
			if lengthVoc == 1 {
				lengthVoc = 2
			}
			for _, wordFromVoc := range vocabulary[lengthVoc] {
				min = distanceSearch(&word, &wordFromVoc, min)
				if min == 0 {
					break loop
				}
			}
		} else {
			if shiftLeft > 0 {
				for _, wordFromVoc := range vocabulary[shiftLeft] {
					min = distanceSearch(&word, &wordFromVoc, min)
					if min == 0 {
						break loop
					}
				}
			}
			if shiftRight <= maxLength {
				for _, wordFromVoc := range vocabulary[shiftRight] {
					min = distanceSearch(&word, &wordFromVoc, min)
					if min == 0 {
						break loop
					}
				}
			}
		}
		shiftLeft--
		shiftRight++
	}
	cache[word] = min
	out <- min
}

func distanceSearch(word *string, wordFromVoc *string, min int) int {
	distance := levenshtein.Dist(*word, *wordFromVoc)
	if distance == 0 {
		return 0
	}
	if distance < min {
		return distance
	}
	return min
}

package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"word/levenshtein"
	"strconv"
	"time"
)

const vacabularyName = "vocabulary.txt"
const buffer = 100

var vocabulary []string

func main() {
	start := time.Now()
	filename := os.Args[1]
	buf := bytes.NewBuffer(nil)
	file, err := os.Open(filename)
	defer file.Close()
	if err != nil {
		panic(fmt.Sprintf("File with name '%s' not found", filename))
	}
	io.Copy(buf, file)

	str := string(buf.Bytes())
	if len(str) == 0 {
		panic(fmt.Sprintf("Empty file!"))
	}
	before := len(str)
	newLen := 0
	for before != newLen {
		before = len(str)
		str = strings.Replace(str, "\n", "", -1)
		str = strings.Replace(str, "  ", " ", -1)
		newLen = len(str)
	}
	words := strings.Split(str, " ")
	fmt.Println(words)
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
func prepareVocabulary() (words []string, err error) {
	buf := bytes.NewBuffer(nil)
	file, err := os.Open(vacabularyName)
	defer file.Close()
	if err != nil {
		return
	}
	io.Copy(buf, file)
	str := strings.TrimSpace(string(buf.Bytes()))
	words = strings.Split(str, "\n")
	return
}

func getDistance(in <-chan string, out chan<- int) {
	word := strings.ToUpper(<-in)
	min := levenshtein.Dist(word, vocabulary[0])
	distance := min
	for _, wordFromVoc := range vocabulary {
		distance = levenshtein.Dist(word, wordFromVoc)
		if distance == 0 {
			min = 0
			break
		}
		if distance < min {
			min = distance
		}
	}
	out <- min
}

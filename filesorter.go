// Please write an algorithm in Golang or Rust,
// utilizing concurrency, that sorts the contents of,
// /n delimited, txt files into nested alphanumeric files
// at a specified directory. If any file reaches a threshold size,
// create a folder with that index and sort by the subsequent
// character into subfiles. Internally sort all files alphanumerically.
// Finally, determine if an input file has already been sorted.

// Questions:
// (1) does "/n delimited" mean "files delimited by \n" (typo? as in
// new line?) or "n files that are delimited"?
// If the latter, delimited by what?
// (2) Are there any constraints on the txt files? I.e., maximum
// # of characters or words, the type of delimiter,
// which kinds of characters, how many files, etc.
// (3) Clarify "sort by the subsequent character into subfiles"?
// I'm unclear on how you want me to "nest" the files
// (4) Possible example output / input?

package main

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"time"
)

// this struct will be used for concurrent communication
type Result struct {
	fileName string
	sorted   bool
	workerID int
}

// prints the slice "words" of words
func printWords(words []string) {
	n := len(words)

	for i := 0; i < n; i++ {
		fmt.Println(words[i])
	}
}

// returns true if the slice "words" is sorted; else returns false
func isSliceSorted(words []string) bool {
	return sort.SliceIsSorted(words, func(p, q int) bool {
		return words[p] < words[q]
	})
}

// returns true if the file "fileName" is sorted; else returns false
func isFileSorted(fileName string) bool {
	return isSliceSorted(scanWordsToSlice(fileName))
}

// randomly shuffles the elements of "words"
func Shuffle(words []string) {

	// set the seed for pseudorandomness
	rand.Seed(time.Now().UnixNano())

	// this uses an anynomous "swap" function
	rand.Shuffle(len(words), func(i, j int) {
		words[i], words[j] = words[j], words[i]
	})
}

// generates a psuedorandom integer in [min, max)
func RNG(min, max int) int {

	// set the seed for pseudorandomness
	rand.Seed(time.Now().UnixNano())

	return (rand.Intn(max-min) + min)
}

// returns a slice containing the words from the file "fileName"
func scanWordsToSlice(fileName string) []string {

	// this slice will contain all the words from a file
	var words []string

	// open the file
	f, err := os.Open(fileName)

	// make sure the file opened correctly
	if err != nil {
		log.Fatal(err)
	}

	// close the file when we're done with it
	defer f.Close()

	// create the scanner to scan the file
	scanner := bufio.NewScanner(f)

	// scan all words into a slice
	for i := 0; scanner.Scan(); i++ {
		words = append(words, scanner.Text())
	}

	// make sure the file was scanned correctly
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	// return the slice of words
	return words
}

// overwrites the contents of fileName with the words in this slice
// if fileName doesn't exist yet, this function will create it
func writeWordsToFile(words []string, fileName string) {
	// truncate / create the txt file
	f, err := os.Create(fileName)

	// make sure the file was truncated / created correctly
	if err != nil {
		log.Fatal(err)
	}

	// close the file when we're done with it
	defer f.Close()

	// write each word to the new file, delimited by "\n"
	for _, word := range words {

		// write word to file, then a line break
		_, err := f.WriteString(word + "\n")

		// make sure the word was written to the file correctly
		if err != nil {
			log.Fatal(err)
		}
	}
}

// creates a new text file containing the same words as fileName,
// named newFileName, but the words are sorted
func createFileWithSortedWords(newFileName, fileName string) {

	// make a slice with the words from this file
	words := scanWordsToSlice(fileName)

	// sort the words
	sort.Strings(words)

	// create and populate this new txt file
	writeWordsToFile(words, newFileName)
}

// creates a new text file containing the same words as fileName,
// named newFileName, but the words are randomly shuffled
func createFileWithShuffledWords(newFileName, fileName string) {

	// make a slice with the words from this file
	words := scanWordsToSlice(fileName)

	// shuffle the words
	Shuffle(words)

	// create and populate this new txt file
	writeWordsToFile(words, newFileName)
}

// sorts the words in a txt file, delimited by "\n"
func sortWordsInFile(fileName string) {

	// make a slice with the words from this file
	words := scanWordsToSlice(fileName)

	// sort the words
	sort.Strings(words)

	// overwrite the file with sorted words
	writeWordsToFile(words, fileName)
}

// shuffles the words in a txt file, delimited by "\n"
func shuffleWordsInFile(fileName string) {

	// make a slice with the words from this file
	words := scanWordsToSlice(fileName)

	// shuffle the words
	Shuffle(words)

	// overwrite the file with shuffled words
	writeWordsToFile(words, fileName)
}

// creates k new txt files in a folder, each containing a random
// selection of words from the txt file fileName. there will be
// a random number of words, and in a random order
func createNewFiles(k int, fileName string) {

	// make a slice with the words from this file
	words := scanWordsToSlice(fileName)
	N := len(words)

	for i := 1; i <= k; i++ {

		// n random words; n is in [0,N]
		n := RNG(0, N+1)

		// randomize the slice
		Shuffle(words)

		// create this new file
		writeWordsToFile(words[0:n], "wordfiles/"+
			strconv.Itoa(i)+fileName)
	}
}

// returns a slice of strings containing the names of all
// of the .txt files at the specified directory
func fetchTextFileNames(directory string) []string {

	// a slice containing all the desired filenames
	fileNames, err := filepath.Glob(directory + "*.txt")

	// make sure we aquired the slice correctly
	if err != nil {
		log.Fatal(err)
	}

	// return the slice
	return fileNames
}

// the concurrent worker function
func worker(jobs <-chan string, results chan<- Result, ID int) {

	// each worker uses this for loop when there is a job
	// available to do, assuming they aren't already
	// working on a different job
	for fileName := range jobs {
		// here the worker sort the words
		sortWordsInFile(fileName)

		// communicates the fileName back, and confirms
		// if the file was indeed sorted correctly
		var result Result
		result.fileName = fileName
		result.sorted = isFileSorted(fileName)
		result.workerID = ID

		// send message
		results <- result
	}
}

func main() {

	// number of files to create and sort
	N := 1000

	// create the files; unsorted and random
	createNewFiles(N, "words.txt")

	// a slice of strings of all names of txt files in this directory
	fileNames := fetchTextFileNames("/home/danmarino900/go_workspace/wordfiles/")

	// the channels over which our workers will recieve their jobs,
	// and communicate their results (all done concurrently)
	jobs := make(chan string, N)
	results := make(chan Result, N)

	// workers ready!
	for i := 1; i <= 9; i++ {
		go worker(jobs, results, i)
	}

	// put all the jobs in the job channel for the workers
	for _, fileName := range fileNames {
		// send message
		jobs <- fileName
	}
	// the sender closes this channel; no more jobs to send
	close(jobs)

	// collect the results from each job as the workers finish them
	for i := 0; i < N; i++ {
		// receive message
		result := <-results

		// print our findings!
		if result.sorted {
			fmt.Printf("Worker %v: Is %-54s sorted? Yes it is! Great job!\n", result.workerID, result.fileName)
		} else {
			fmt.Printf("Oh no! Worker %v: %-54s is not sorted! Something went wrong =(\n", result.workerID, result.fileName)
			os.Exit(result.workerID)
		}
	}
}

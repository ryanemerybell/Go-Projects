// Ryan Bell - Genity Problem Statement

// Please write an algorithm in Golang or Rust,
// utilizing concurrency, that sorts the contents of,
// /n delimited, txt files into nested alphanumeric files
// at a specified directory. If any file reaches a threshold size,
// create a folder with that index and sort by the subsequent
// character into subfiles. Internally sort all files alphanumerically.
// Finally, determine if an input file has already been sorted.

package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

var wg sync.WaitGroup
var thresholdSize int = 1000

// returns a slice containing the contents of the input file
func scanContentsToSlice(fileName string) []string {
	var contents []string

	// open the file; close it after use
	f, err := os.Open(fileName)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	// create the scanner and scan all contents into a slice
	scanner := bufio.NewScanner(f)
	for i := 0; scanner.Scan(); i++ {
		contents = append(contents, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return contents
}

// generates a psuedorandom integer in [min, max)
func RNG(min, max int) int {
	// set the seed for pseudorandomness
	rand.Seed(time.Now().UnixNano())

	return (rand.Intn(max-min) + min)
}

// randomly shuffles the elements of the input slice
func Shuffle(contents []string) {
	rand.Seed(time.Now().UnixNano())

	// this uses an anynomous "swap" function
	rand.Shuffle(len(contents), func(i, j int) {
		contents[i], contents[j] = contents[j], contents[i]
	})
}

// overwrites the contents of fileName with the contents in this slice
// if fileName doesn't exist yet, this function will create it
func overwriteContentsToFile(contents []string, fileName string) {
	// truncates or creates the file; close file when done
	f, err := os.Create(fileName)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	// write each content to the new file, delimited by "\n"
	for _, content := range contents {
		if _, err := f.WriteString(content + "\n"); err != nil {
			log.Fatal(err)
		}
	}

}

// creates k new text files in the input directory, each containing
// a random selection of contents from the input txt file
func createNewFiles(k int, directory, fileName string) {
	contents := scanContentsToSlice(fileName)
	N := len(contents)

	for i := 1; i <= k; i++ {
		n := RNG(0, N+1)
		Shuffle(contents)
		overwriteContentsToFile(contents[0:n], directory+"/"+
			strconv.Itoa(i)+fileName)
	}
}

// returns a slice of strings containing the names of all
// of the txt files at the specified directory
func fetchTextFileNames(directory string) []string {
	fileNames, err := filepath.Glob(directory + "*.txt")
	if err != nil {
		log.Fatal(err)
	}

	return fileNames
}

// check if the input slice is sorted
func isSliceSorted(words []string) bool {
	return sort.SliceIsSorted(words, func(p, q int) bool {
		return words[p] < words[q]
	})
}

// checks if the input file is sorted
func isFileSorted(fileName string) bool {
	return isSliceSorted(scanContentsToSlice(fileName))
}

// creates a map that splits up the contents slice into new slices
// of strings which are organized by their first depth characters
func sortContentsIntoMap(contents []string, depth int) map[string][]string {
	characters := make(map[string][]string)

	for _, content := range contents {
		// lines of contents that are too short get padded with " "
		if length := utf8.RuneCountInString(content); length < depth {
			content = fmt.Sprintf("%-*v", depth, content)
		}
		s := content[0:depth]

		// have we seen these initial characters before?
		// if no, put it in the map. in either case,
		// add the content in the appropriate slice
		if _, ok := characters[s]; ok {
			characters[s] = append(characters[s], content)
		} else {
			characters[s] = make([]string, 0)
			characters[s] = append(characters[s], content)
		}
	}

	return characters
}

// writes a slice of strings to a file (sorted), or if the slice is
// too long, creates a new directory and calls itself recursively
func contentSorterRec(contents []string, directory string, depth int) {
	if contents == nil {
		log.Fatal(errors.New("Nothing to write into file!"))
	}

	// this will determine the name of the new file or folder
	firstContent := contents[0]
	newName := directory + firstContent[0:depth]

	// if we reach the thresholdSize, sort by the subsequent character.
	// otherwise, sort the contents and write them to a file
	if len(contents) >= thresholdSize {
		os.Mkdir(newName, os.ModePerm)
		characters := sortContentsIntoMap(contents, depth+1)
		for _, newContents := range characters {
			contentSorterRec(newContents, newName+"/", depth+1)
		}
	} else {
		sort.Strings(contents)
		overwriteContentsToFile(contents, newName+".txt")
	}
}

// this is the concurrent worker function,
// each of which sorts contents recursively
func worker(jobs <-chan []string, directory string) {
	for contents := range jobs {
		contentSorterRec(contents, directory+"/", 1)
	}

	wg.Done()
}

// this function reads all contents from all files in inputDirectory
// into memory at once. this trades space efficiency for speed.
// then, it (concurrently) writes the contents from those files
// into outputDirectory, sorted in the desired fashion
func contentSorter(outputDirectory, inputDirectory string) {
	// we begin sorting all contents by their first character
	firstCharacter := make(map[string][]string)

	// open all files sequentially
	for _, fileName := range fetchTextFileNames(inputDirectory + "/") {
		f, err := os.Open(fileName)
		if err != nil {
			log.Fatal(err)
		}

		// determine if input file is already sorted
		if isFileSorted(fileName) {
			fmt.Println("Input File", fileName, "is already sorted!")
		}

		// create the scanner and scan the file
		scanner := bufio.NewScanner(f)
		for i := 0; scanner.Scan(); i++ {
			content := scanner.Text()
			if strings.Compare(content, "") == 0 {
				continue
			}
			c := string(content[0])

			// have we seen this character before?
			// if no, put it in the map. in either case,
			// add the content in the appropriate slice
			if _, ok := firstCharacter[c]; ok {
				firstCharacter[c] = append(firstCharacter[c], content)
			} else {
				firstCharacter[c] = make([]string, 0)
				firstCharacter[c] = append(firstCharacter[c], content)
			}
		}

		if err := scanner.Err(); err != nil {
			log.Fatal(err)
		}
		f.Close()
	}

	// this is the concurrency block.
	// we have a buffered channel of contents to sort,
	// and "numWorkers" workers ready to do so concurrently.
	// the waitgroup waits until all workers are done, and
	// all workers will be done after there are no more jobs.
	jobs := make(chan []string, len(firstCharacter))
	numWorkers := 8
	wg.Add(numWorkers)
	for i := 1; i <= numWorkers; i++ {
		go worker(jobs, outputDirectory)
	}
	for _, contents := range firstCharacter {
		jobs <- contents
	}
	close(jobs)
}

func main() {
	inputDirectory := "namefiles"
	outputDirectory := "sortednames"
	numInputFiles := 100
	defer wg.Wait()

	// create new random input files
	os.RemoveAll(inputDirectory)
	os.Mkdir(inputDirectory, os.ModePerm)
	createNewFiles(numInputFiles, inputDirectory, "names.txt")

	// sort the contents of those input files as desired
	os.RemoveAll(outputDirectory)
	os.Mkdir(outputDirectory, os.ModePerm)
	contentSorter(outputDirectory, inputDirectory)
}

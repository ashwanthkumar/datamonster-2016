package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"container/heap"

	"github.com/ashwanthkumar/datamonster_2016/hset"
	"github.com/ashwanthkumar/datamonster_2016/ngram"
)

const (
	// MaxNgrams - maximum ngrams to create for titles
	MaxNgrams = 2
	// MaxItemsInBag - maximum number of most occuring items to keep in each bag
	MaxItemsInBag = 100
)

// BrandTokens - Contains the Bag of words of all brands as branId -> sets.Set
var BrandTokens map[int]hset.MapSetBasedHeap

// WordToBrandMap - Inverted index of frequency to array of brands
var WordToBrandMap map[string][]int

func init() {
	BrandTokens = make(map[int]hset.MapSetBasedHeap)
	WordToBrandMap = make(map[string][]int)
}

func main() {
	// readAndTrainDataset("datasets/classification_train.tsv")
	dataset := readAndTrainDataset("datasets/xaa")
	// dataset := readAndTrainDataset("datasets/xaa_small")
	predictFromDataset(dataset)
	// predictFrom(os.Stdin)
	// file, _ := os.Open("/tmp/train-accuracy")
	// predictFrom(file)
	// fmt.Printf("%v\n", BrandTokens)
}

func predictFromDataset(dataset []*TrainDataset) {
	// dumpIV()
	for _, item := range dataset {
		brandID, score := predictBrand(item.Title)
		fmt.Printf("%d\t%d\t%v\t%v\n", item.Brand, brandID, score, brandID == item.Brand)
	}
}

func dumpIV() {
	for word, brandIds := range WordToBrandMap {
		fmt.Printf("%s - %v\n", word, brandIds)
	}
}

// func predictFrom(input io.Reader) {
// 	scanner := bufio.NewScanner(input)
// 	for scanner.Scan() {
// 		brandID, score := predictBrand(scanner.Text())
// 		fmt.Printf("%d\t%v\n", brandID, score)
// 		// fmt.Printf("Most frequent word in the brand(%d) is %s\n", brandID, BrandTokens[brandID].MaxOccuringItem())
// 	}
// 	if err := scanner.Err(); err != nil {
// 		fmt.Fprintln(os.Stderr, "reading standard input:", err)
// 	}
// }

func predictBrand(input string) (int, float64) {
	bagOfWordsForInput := computeBagOfWordsFor(input)
	brandFrequency := hset.Empty()
	var tokensMatched = 0
	for _, word := range bagOfWordsForInput.Values() {
		brandIds, present := WordToBrandMap[word]
		if present {
			tokensMatched++
			for _, brandID := range brandIds {
				brandFrequency.Add(strconv.Itoa(brandID))
			}
		}
	}

	sort.Stable(&brandFrequency)

	var actualBrandID int
	var score float64
	if brandFrequency.Len() > 0 {
		actualBrandID = toInt(brandFrequency.MaxOccuringItem())
		score = float64(tokensMatched) / float64(bagOfWordsForInput.Len())
	} else {
		actualBrandID = -1
		score = 0.0
	}
	return actualBrandID, score
}

func computeBagOfWordsFor(input string) hset.MapSetBasedHeap {
	var ngrams []string
	for n := 1; n <= MaxNgrams; n++ {
		tokens, err := ngram.Tokenize(n, input)
		if err != nil {
			fmt.Printf("Input Line - %s\n", input)
			panic(err)
		}
		ngrams = append(ngrams, tokens...)
	}
	return hset.FromSlice(ngrams)
}

func readAndTrainDataset(input string) []*TrainDataset {
	file, err := os.Open(input)
	defer file.Close()
	if err != nil {
		panic(err)
	}

	var dataset []*TrainDataset
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		item := newDataset(scanner.Text())
		dataset = append(dataset, item)

		bagOfWords, exists := BrandTokens[item.Brand]
		if !exists {
			bagOfWords = hset.Empty()
		}
		newBagOfWords := bagOfWords.Union(computeBagOfWordsFor(item.Title))
		BrandTokens[item.Brand] = newBagOfWords
		heap.Init(&newBagOfWords)
		for newBagOfWords.Len() > MaxItemsInBag {
			heap.Pop(&newBagOfWords)
		}

		length := len(dataset)
		if length > 0 && length%1000 == 0 {
			// if length > 0 {
			fmt.Printf("Processed %d lines so far\n", length)
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading from "+input+": ", err)
	}

	fmt.Println("Building Inverted Index")
	// Build the Inverted index on words
	for brandID, bagOfWords := range BrandTokens {
		for _, word := range bagOfWords.Values() {
			brandIds, _ := WordToBrandMap[word]
			brandIds = append(brandIds, brandID)
			WordToBrandMap[word] = brandIds
		}
	}
	fmt.Println("Inverted Index Built")

	return dataset
}

// TrainDataset - Dataset format as in the input file
type TrainDataset struct {
	Title    string
	Brand    int
	Category int
}

func newDataset(line string) *TrainDataset {
	tokens := strings.Split(line, "\t")
	return &TrainDataset{
		Title:    tokens[0],
		Brand:    toInt(tokens[1]),
		Category: toInt(tokens[2]),
	}
}

func toInt(input string) int {
	data, _ := strconv.ParseInt(input, 10, 32)
	return int(data)
}

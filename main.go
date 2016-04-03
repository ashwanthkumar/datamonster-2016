package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"

	"container/heap"

	"github.com/ashwanthkumar/datamonster_2016/hset"
	myngram "github.com/ashwanthkumar/datamonster_2016/ngram"
	ngram "github.com/lestrrat/go-ngram"
)

const (
	// MaxNgrams - maximum ngrams to create for titles
	MaxNgrams = 2
	// MaxItemsInBag - maximum number of most occuring items to keep in each bag
	MaxItemsInBag = 5000
	// BrandCountCutOff - Words lesser than this many occurances are ignored
	BrandCountCutOff = 100
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
	trainDataset := os.Args[1]
	testDataset := os.Args[2]
	brandDistributionFile := os.Args[3]
	outputLocation := os.Args[4]

	fmt.Printf("Train Dataset = %s\n", trainDataset)
	fmt.Printf("Test Dataset = %s\n", testDataset)
	fmt.Printf("BrandDistribution = %s\n", brandDistributionFile)
	fmt.Printf("Output Location = %s\n", outputLocation)

	Workers := 1000
	jobsChannel := make(chan Input)
	resultsChannel := make(chan Output)

	var wg sync.WaitGroup

	for count := 0; count < Workers; count++ {
		go brandPredictWorker(wg, jobsChannel, resultsChannel)
	}
	outputFile, err := os.Create(outputLocation)
	if err != nil {
		panic(err)
	}
	defer outputFile.Close()
	go outputWriter(outputFile, resultsChannel)

	// PREDICTION TYPE1 STARTS
	// dataset := readAndTrainDataset(trainDataset, brandDistributionFile)
	// predictFromDataset(dataset, jobsChannel)
	// PREDICTION TYPE1 ENDS

	// PREDICTION TYPE2 STARTS
	readAndTrainDataset(trainDataset, brandDistributionFile)
	file, _ := os.Open(testDataset)
	defer file.Close()
	predictFrom(file, jobsChannel)
	// PREDICTION TYPE2 ENDS

	for count := 0; count < Workers; count++ {
		jobsChannel <- Input{Seq: -1}
	}
	wg.Wait()
	close(jobsChannel)
	close(resultsChannel)
}

func predictFromDataset(dataset []*TrainDataset, jobsChannel chan Input) {
	var seq = 0
	for _, item := range dataset {
		seq++
		if seq > 0 && seq%10000 == 0 {
			fmt.Printf("[DEBUG] Processed %d product titles so far\n", seq)
		}
		input := Input{
			Title:         item.Title,
			Seq:           seq,
			ExpectedBrand: item.Brand,
		}
		jobsChannel <- input
		// brandID, score := predictBrand(item.Title)
		// 	fmt.Printf("%d\t%d\t%v\t%v\n", item.Brand, brandID, score, brandID == item.Brand)
	}
}

type Output struct {
	BrandId       int
	Seq           int
	ExpectedBrand int
}

type Input struct {
	Title         string
	Seq           int
	ExpectedBrand int
}

func predictFrom(input io.Reader, jobsChannel chan Input) {
	scanner := bufio.NewScanner(input)
	var seq = 0
	for scanner.Scan() {
		seq++
		if seq > 0 && seq%10000 == 0 {
			fmt.Printf("[DEBUG] Processed %d product titles so far\n", seq)
		}
		// item := newDataset(scanner.Text())
		item := newTestDataset(scanner.Text())
		input := Input{
			Title:         item.Title,
			Seq:           seq,
			ExpectedBrand: item.Brand,
		}
		jobsChannel <- input
	}

	fmt.Printf("[DEBUG] Processed %d product titles in total\n", seq)
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}
}

func outputWriter(outputFile *os.File, resultsChannel chan Output) {
	var running = true
	for running {
		select {
		case output := <-resultsChannel:
			if output.Seq == -1 {
				running = false
			} else {
				// fmt.Printf("%d\t%d\n", output.BrandId, output.Seq)
				fmt.Fprintf(outputFile, "%d\t%d\t%v\t%v\n", output.BrandId, output.ExpectedBrand, output.Seq, output.BrandId == output.ExpectedBrand)
				// fmt.Printf("%d\t%d\t%v\t%v\n", output.BrandId, output.ExpectedBrand, output.Seq, output.BrandId == output.ExpectedBrand)
			}
		}
	}
}

func brandPredictWorker(wg sync.WaitGroup, jobs <-chan Input, output chan<- Output) {
	wg.Add(1)
	var running = true
	for running {
		select {
		case job := <-jobs:
			if job.Seq == -1 {
				running = false
				wg.Done()
			} else {
				brandID, _ := predictBrand(job.Title)
				// fmt.Printf("%d\t%v\n", brandID, score)
				op := Output{
					BrandId:       brandID,
					Seq:           job.Seq,
					ExpectedBrand: job.ExpectedBrand,
				}
				output <- op
			}
		}
	}
}

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
	return createBagOfWords(input)
	// return createTokenBasedNGrams(input)
}

func createTokenBasedNGrams(input string) hset.MapSetBasedHeap {
	var ngrams []string

	// Clean the stop words from the input
	inputTokens := hset.FromSlice(strings.Split(input, " "))
	for _, stopWord := range StopWords.Values() {
		inputTokens.Remove(stopWord)
	}
	cleanedInput := strings.Join(inputTokens.Values(), " ")

	for n := 1; n <= MaxNgrams; n++ {
		if len(cleanedInput) > n {
			tokens := ngram.NewTokenize(n, cleanedInput).Tokens()
			for _, token := range tokens {
				ngrams = append(ngrams, token.String())
			}
		}
	}

	bagOfWords := hset.FromSlice(ngrams)
	return bagOfWords
}

func createBagOfWords(input string) hset.MapSetBasedHeap {
	var ngrams []string
	for n := 1; n <= MaxNgrams; n++ {
		tokens, err := myngram.Tokenize(n, input)
		// tokens, err := ngram.NewTokenize(n, input).Tokens
		if err != nil {
			fmt.Printf("Input Line - %s\n", input)
			panic(err)
		}

		tokenSize := len(tokens)
		if tokenSize > 3 {
			start, end := tokens[:3], tokens[tokenSize-3:]
			ngrams = append(ngrams, start...)
			ngrams = append(ngrams, end...)
		} else {
			ngrams = append(ngrams, tokens...)
		}
	}
	bagOfWords := hset.FromSlice(ngrams)
	// strip the stop words from the list
	for _, stopWord := range StopWords.Values() {
		bagOfWords.Remove(stopWord)
	}

	return bagOfWords
}

func readBrandDistribution(brandDistributionFile string) map[int]int {
	file, err := os.Open(brandDistributionFile)
	defer file.Close()
	if err != nil {
		panic(err)
	}

	var brandCounts = make(map[int]int)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		// count -> brandId
		tokens := strings.Split(line, " ")
		brandCounts[toInt(tokens[1])] = toInt(tokens[0])
	}

	return brandCounts
}

func readAndTrainDataset(input, brandDistributionFile string) []*TrainDataset {
	file, err := os.Open(input)
	defer file.Close()
	if err != nil {
		panic(err)
	}

	var dataset []*TrainDataset
	var brandDistributions = readBrandDistribution(brandDistributionFile)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		item := newDataset(scanner.Text())
		if brandDistributions[item.Brand] < BrandCountCutOff {
			continue
		}

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
		if length > 0 && length%10000 == 0 {
			// if length > 0 && length%1000 == 0 {
			// if length > 0 {
			fmt.Printf("Processed %d lines so far\n", length)
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading from "+input+": ", err)
	}
	length := len(dataset)
	fmt.Printf("Processed %d lines so far\n", length)

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

func newTestDataset(line string) *TrainDataset {
	tokens := strings.Split(line, "\t")
	return &TrainDataset{
		Title:    tokens[0],
		Category: toInt(tokens[1]),
	}
}

func toInt(input string) int {
	data, _ := strconv.ParseInt(input, 10, 32)
	return int(data)
}

package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"sync"
	"time"
)

type mainReader struct {
	words *[]mainRecord
}

type mainRecord struct{
	word   []byte
	counter int
}

func (m *mainReader) contains(element []byte) (bool, int){
	for index, v := range *m.words {
		if bytes.Equal(v.word, element){
			return true, index
		}
		index = index + 1
	}
	return false, 0
}

func( m *mainReader) main_read_from_chan(ch chan []byte) {
	wg := &sync.WaitGroup{}
	mu := &sync.Mutex{}

	for node := range ch {
		wg.Add(1)

		node := node

		go func() {
			state, index := m.contains(node)

			if state{
				mu.Lock()
				(*m.words)[index].counter++
				mu.Unlock()
			} else{
				record := mainRecord{node, 1}
				mu.Lock()
				*m.words = append(*m.words, record)
				mu.Unlock()
			}

			wg.Done()
		}()
	}
	wg.Wait()
}

func(m *mainReader) get20MostFrequentlyUsedWords(){
	sort.Slice(*m.words, func(i, j int) bool {
		return (*m.words)[i].counter > (*m.words)[j].counter
	})
}
//I created two structs with the methods to write and read from the shared channel of []bytes. String is not allowed, so we assume that slice of bytes is a word.
//Writer and reader works in different goroutines.

//Writer goes through the text and collects bytes to buffer until it reaches the space(which means the end of the word),
//or reaches the symbol which is not a letter(in that case we check do we already have a word in our buffer).
//After this it sends the content of the buffer to channel, and continues to go through the text

//At the same time reader listens to the channel. Whenever it gets the word ([]byte) it checks does it have the same word inside the slice of already written words.
//If it does, it increases the counter of this record, if no, it appends the record to it.
//After writer and reader both finished working with a channel, reader goes through the slice of words to determine the most 20 frequent words. I did this with an assumption that,
//any sorting between approximately 74k elements while reading, or quick sorting 10k elements after reading,
//will lead to way more comparisons than going through the slice and just retrieving 20 elements with the largest counter(10k elements in slice, 74k words in text)
//P.S Using one goroutine showed better execution time, idk why, but I decided to go with two goroutines version.
//I used approx 5 hours to code version with no goroutines, and one extra hour to code version that u see

func solution(){
	file, err := os.Open("output.txt")
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		if err = file.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	readingBuf := make([]byte, 1)

	words := make([]mainRecord, 0)
	reader := mainReader{words: &words}

	writingBuf := make([]byte, 0)

	ch := make(chan []byte)

	m := bufio.NewReader(file)
	go func() {
		for {
			//reading file's letters one by one
			n, err := m.Read(readingBuf)
			if n > 0 {
				byteVal := readingBuf[0]
				if byteVal >= 65 && byteVal <= 90 { //if symbol is uppercase letter
					byteVal = byteVal + 32
					writingBuf = append(writingBuf, byteVal)
					//writing to temporary buffer
				} else if byteVal >= 97 && byteVal <= 122 { //if symbol is lowercase letter
					writingBuf = append(writingBuf, byteVal)
					//writing to temporary buffer
				} else if byteVal == 32 && len(writingBuf) != 0 { //if symbol is [space], and we have letters in our buffer
					ch <- writingBuf
					writingBuf = nil

				} else if ((byteVal > 122 || byteVal < 65) || (byteVal > 90 && byteVal < 97)) && len(writingBuf) != 0 { //if symbol is any other than letter or space, and we have letters in our buffer
					ch <- writingBuf
					writingBuf = nil

				} else {
					continue
				}
			}

			if err == io.EOF {
				ch <- writingBuf
				writingBuf = nil //send temporary buffer content to channel, empty the temporary buffer
				break
			}
		}

		close(ch) //close channel, so our that our reader will stop working after there are no elements left, in other case reader will cause deadlock
	}()

	reader.main_read_from_chan(ch) //reading from channel in range of elements in channel

	reader.get20MostFrequentlyUsedWords() //getting 20 // most frequent words, and write it to rating slice

}
func main() {

	start := time.Now()
	solution()
	fmt.Printf("Process took %s\n", time.Since(start))
}

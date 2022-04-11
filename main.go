package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const HEADER_ANIM_DELAY = time.Millisecond * 200

const URL_API = "https://translate.google.com/translate_a/single?client=at&dt=t&dt=ld&dt=qca&dt=rm&dt=bd&dj=1&ie=UTF-8&oe=UTF-8&inputm=2&otf=2&iid=1dd3b944-fa62-4b55-b330-74909a99969e"

var clear map[string]func()

func init() {
	clear = make(map[string]func())
	clear["linux"] = func() {
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
	clear["windows"] = func() {
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
}

func ClearCMD() {
	clearFunc, ok := clear[runtime.GOOS]
	if ok {
		clearFunc()
	}
}

type Output struct {
	Sentences []Sentence `json:"sentences"`
}
type Sentence struct {
	Trans string `json:"trans"`
}

func TranslateRequestAPI(origin string, target string, s string, ch chan string) {
	data := url.Values{}

	// origin language code
	data.Set("sl", origin)
	// target language code
	data.Set("tl", target)
	// sentence to translate
	data.Set("q", s)

	// create client object
	client := &http.Client{}
	r, err := http.NewRequest("POST", URL_API, strings.NewReader(data.Encode())) // URL-encoded payload
	if err != nil {
		log.Fatal(err)
	}

	// add header for x-www-form-urlencoded
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

	// do defined request
	res, err := client.Do(r)
	if err != nil {
		log.Fatal(err)
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	result := Output{}

	if err := json.Unmarshal(body, &result); err != nil {
		log.Printf("error: %s", err.Error())
	}

	ch <- result.Sentences[0].Trans
}

func NewLine(c int) {
	for j := 0; j < c; j++ {
		fmt.Println()
	}
}

func TabLine(c int) {
	for j := 0; j < c; j++ {
		fmt.Printf("\t")
	}
}

func main() {
	for i := 0; i < 5; i++ {
		ClearCMD()
		NewLine(i)
		TabLine(i)
		fmt.Println("+-----------------------------+")
		TabLine(i)
		fmt.Println("|    SIMPLE TRANSLATION APP   |")
		TabLine(i)
		fmt.Println("+-----------------------------+")
		time.Sleep(HEADER_ANIM_DELAY)
	}

	fmt.Println("\n\n\t\t\t      This app using Google Translate API")
	fmt.Println("\t\t\t  For supported language and its code see below:")
	fmt.Println("\t\t\thttps://cloud.google.com/translate/docs/languages")
	fmt.Println("\n\n\t   <<-Note: Please resize the [terminal] / [CMD] if the symbols are messed up->>")

	var prompt string

	fmt.Println("\n\nEnter [any] to continue and [q] for exit...")
	fmt.Scanln(&prompt)
	prompt = strings.TrimSpace(prompt)

	if prompt == "q" {
		fmt.Println("Bye-bye!")
		os.Exit(0)
	}

	for {
		for i := 5; i > 0; i-- {
			ClearCMD()
			NewLine(i)
			TabLine(4)
			fmt.Println("+-----------------------------+")
			TabLine(4)
			fmt.Println("|    SIMPLE TRANSLATION APP   |")
			TabLine(4)
			fmt.Println("+-----------------------------+")
			time.Sleep(HEADER_ANIM_DELAY)
		}

		var sentence string
		var origin string
		var target string

		inputReader := bufio.NewReader(os.Stdin)

		fmt.Printf("\n> enter the word(s) you want to translate: ")
		sentence, _ = inputReader.ReadString('\n')
		sentence = strings.Trim(sentence, "\n")

		fmt.Printf("\n> enter language code origin: ")
		fmt.Scanln(&origin)

		fmt.Printf("> enter language code target: ")
		fmt.Scanln(&target)

		ch := make(chan string)

		go TranslateRequestAPI(origin, target, sentence, ch)

		loading := func() {
			isDone := false
			fmt.Printf("\n\n\nTRANSLATING [")
		LoadingLoop:
			for i := 0; i < 10; i++ {
				fmt.Printf("|")
				time.Sleep(time.Millisecond * 200)
				if i >= 4 {
					select {
					case trans := <-ch:
						isDone = true
						fmt.Printf("|||||||||||||]  FINISHED")
						time.Sleep(time.Millisecond * 500)

						ClearCMD()
						TabLine(5)
						fmt.Println("+------------------+")
						TabLine(5)
						fmt.Println("|      RESULT      |")
						TabLine(5)
						fmt.Printf("+------------------+\n\n")
						TabLine(1)
						fmt.Printf("> [%s]: %s\n", origin, sentence)
						TabLine(1)
						fmt.Printf("> [%s]: %s\n", target, trans)

						break LoadingLoop
					default:
						continue
					}
				}
			}
			if !isDone {
				fmt.Printf("]")
				time.Sleep(time.Millisecond * 100)
				fmt.Println("Failed contacting Google, app exit..")
				os.Exit(1)
			}
		}

		loading()

		fmt.Println("\n\nEnter [any] to translate again and [q] for quit...")
		prompt, _ := inputReader.ReadString('\n')
		prompt = strings.Trim(prompt, "\n")
		if prompt == "q" {
			fmt.Println("Bye-bye!")
			os.Exit(0)
		}
	}

}

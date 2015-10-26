/*

mybot - Illustrative Slack bot in Go

Copyright (c) 2015 RapidLoop

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/

package main

import (
	"encoding/csv"
	"fmt"
	"io/ioutil"
    "log"
	"net/http"
	"os"
	"strings"
    "bytes"
    "encoding/json"
	"github.com/mvdan/xurls"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "usage: mybot slack-bot-token\n")
		os.Exit(1)
	}

	// start a websocket-based Real Time API session
	ws, id := slackConnect(os.Args[1])
	fmt.Println("mybot ready, ^C exits")
	fmt.Println(fmt.Sprintf("id is %s", id))

	for {
		// read each incoming message
		m, err := getMessage(ws)
		if err != nil {
			log.Fatal(err)
		}

		// see if we're mentioned
		if m.Type == "message" && strings.HasPrefix(m.Text, "<@"+id+">") {
			// if so try to parse if
			parts := strings.Fields(m.Text)
			if len(parts) == 3 && parts[1] == "stock" {
				// looks good, get the quote and reply with the result
				go func(m Message) {
					m.Text = getQuote(parts[2])
					postMessage(ws, m)
				}(m)
				// NOTE: the Message object is copied, this is intentional
			} else {
				// huh?
				m.Text = fmt.Sprintf("sorry, that does not compute\n")
				postMessage(ws, m)
			}
		}

		if m.Type == "message" {
			urls := []string(xurls.Relaxed.FindAllString(m.Text, -1))
			if len(urls) > 0 {
				for _, url := range urls {
					if strings.Contains(url, "|") {
						//x := Message{}
						divvy := strings.SplitN(url, "|", 2)
						fmt.Println("keep " + divvy[0])
                        /*						
                        if strings.EqualFold(id, m.User) {
							x = Message{Type: "message", Channel: m.Channel, Text: "this was from me, so I won't post it and cause a loop."}
						} else {
							x = Message{Type: "message", Channel: m.Channel, Text: "saving an DIVIDED url..." + divvy[0]}
						}
						fmt.Println(x)
						postMessage(ws, x)
                        */
                        postUrl(m.Text, divvy[0])
					} else {
						//x := Message{}
						fmt.Println("no need to divide this one up:" + url)
						/*
                        if strings.EqualFold(id, m.User) {
							x = Message{Type: "message", Channel: m.Channel, Text: "this was from me, so I won't post it and cause a loop."}
						} else {
							x = Message{Type: "message", Channel: m.Channel, Text: "saving an undivided url..." + url}
						}
						fmt.Println(x)
						postMessage(ws, x)
                        */
                        postUrl(m.Text, url)
					}
				}
			}
		}

		/*if m.Type == "message" && strings.Contains(m.Text, "http") {
			username, err := getUserName(os.Args[1], m.User)
			if err != nil {
				log.Fatal(err)
			}

			m.Text = "I guess I should save that, eh " + m.User + ", AKA " + username + "?"
			postMessage(ws, m)
		}*/
	}
}

// Get the quote via Yahoo. You should replace this method to something
// relevant to your team!
func getQuote(sym string) string {
	sym = strings.ToUpper(sym)
	url := fmt.Sprintf("http://download.finance.yahoo.com/d/quotes.csv?s=%s&f=nsl1op&e=.csv", sym)
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Sprintf("error: %v", err)
	}
	rows, err := csv.NewReader(resp.Body).ReadAll()
	if err != nil {
		return fmt.Sprintf("error: %v", err)
	}
	if len(rows) >= 1 && len(rows[0]) == 5 {
		return fmt.Sprintf("%s (%s) is trading at $%s", rows[0][0], rows[0][1], rows[0][2])
	}
	return fmt.Sprintf("unknown response format (symbol was \"%s\")", sym)
}

type urlPost struct {
    Raw string `json:"raw"`
    Url string `json:"url"`
}

func postUrl(raw string, in_url string) (id int, err error) {
    apiUrl := "http://localhost:3000/links"
    postBody := urlPost{Raw:raw, Url: in_url}
    if err != nil {
        log.Fatal(err)
    }

    pu, err := json.Marshal(postBody)
    client := &http.Client{}
    resp, err := client.Post(apiUrl, "application/json", bytes.NewBufferString(string(pu)))

    if err != nil {
        log.Fatal(err)
        return
    }
    if resp.StatusCode != 200 && resp.StatusCode != 201 {
        fmt.Println("I got to the postUrl function... Here is the error!:")
        err = fmt.Errorf("API request failed with code %d", resp.StatusCode)
        fmt.Println(err)
        fmt.Println(resp)
        return
    }
    body, err := ioutil.ReadAll(resp.Body)
    fmt.Println(body)
    resp.Body.Close()
    if err != nil {
        fmt.Println("I got to the postUrl function... 6")
        return
    }
    fmt.Println("I got to the postUrl function... 7")

    id = 999
    return
}

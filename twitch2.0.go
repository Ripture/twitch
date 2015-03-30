package main

import (
	"./lib"
	"encoding/json"
	"fmt"
	"github.com/nsf/termbox-go"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"
	"unicode/utf8"
)

var text string
var cursx int
var cursy int
var GameList []forms.Games
var StreamerList []forms.Streamers

const coldef = termbox.ColorDefault

func main() {

	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()

	termbox.SetInputMode(termbox.InputEsc)

	args := os.Args[1:]

	if len(args) > 0 {
		if args[0] == "/?" || args[0] == "?" {
			fmt.Printf("\nTwitchGo Usage: twitch -s [streamer name] -g [game name]\n\n")
			return
		}
	}

	Streams, err := GetStreams(525)

	if err != nil {
		return
	}

	// //pick a random stream
	rand.Seed(time.Now().UTC().UnixNano())
	val := rand.Intn(len(Streams.Streams))
	currGame := Streams.Streams[val].Game
	currStreamer := Streams.Streams[val].Channel.Name
	_ = currGame
	_ = currStreamer
	_ = Streams

	ShowMainMenu(currGame, currStreamer, Streams, GameList, StreamerList)
}

func GetStreams(num int) (forms.StreamS, error) {
	var SomeStreams forms.StreamS
	var Streams forms.StreamS
	var err error
	var offset int
	var prog float32
	prog = 0
	progposx := 1

	termbox.Clear(coldef, coldef)
	termbox.SetCursor(0, 0)

	drawInitStreamList()

	baseURL := "https://api.twitch.tv/kraken"

	numt := num / 100 //number of times we need to get max limit (100) GETs
	numm := num % 100 //number to pick up the remainder of limit for GET (will be limit=numm)

	for i := 0; i < numt; i++ {
		offset = 100 * i
		resp, err := http.Get(baseURL + "/streams?limit=100&offset=" + strconv.Itoa(offset))
		defer resp.Body.Close()
		if err != nil {
			fmt.Println(err)
			return Streams, err
		}

		body, err := ioutil.ReadAll(resp.Body)

		err = json.Unmarshal(body, &SomeStreams)
		if err != nil {
			fmt.Println("error:", err)
			return Streams, err
		}

		prog += 100 / float32(num) * 100

		bars := int(((100 / float32(num)) * 100) / 3.44)
		for i := bars; i > 0; i-- {
			putCh(progposx, 2, '▓', coldef, coldef)
			progposx += 1
		}

		termbox.SetCursor(32, 2)
		fmt.Print(int(prog))
		termbox.HideCursor()

		Streams = BuildStreamList(Streams, SomeStreams)
	}

	offset = numt * 100

	if numm > 0 { //if there are any remaining streams to get
		resp, err := http.Get(baseURL + "/streams?limit=" + strconv.Itoa(numm) + "&offset=" + strconv.Itoa(offset))
		if err != nil {
			fmt.Println(err)
			return Streams, err
		}

		body, err := ioutil.ReadAll(resp.Body)

		err = json.Unmarshal(body, &SomeStreams)
		if err != nil {
			fmt.Println("error:", err)
			return Streams, err
		}

		for i := progposx; i < 30; i++ {
			putCh(progposx, 2, '▓', coldef, coldef)
			progposx += 1
		}

		Streams = BuildStreamList(Streams, SomeStreams)
	}

	for progposx < 30 {
		putCh(progposx, 2, '▓', coldef, coldef)
		progposx += 1
	}
	termbox.SetCursor(32, 2)
	fmt.Print(100)
	termbox.HideCursor()

	time.Sleep(time.Millisecond * 750)
	return Streams, err
}

func BuildStreamList(Streams forms.StreamS, SomeStreams forms.StreamS) forms.StreamS {

	var tempGame forms.Games
	var tempStream forms.Streamers
	found := false
	_ = found

	for i := 0; i < len(SomeStreams.Streams); i++ {
		Streams.Streams = append(Streams.Streams, SomeStreams.Streams[i])
		tempGame.Name = SomeStreams.Streams[i].Game
		tempGame.Viewers = SomeStreams.Streams[i].Viewers
		tempStream.Viewers = tempGame.Viewers
		tempStream.Name = SomeStreams.Streams[i].Channel.Name
		tempStream.Game = SomeStreams.Streams[i].Game
		StreamerList = append(StreamerList, tempStream)

		for j := 0; j < len(GameList); j++ {
			if GameList[j].Name == tempGame.Name {
				GameList[j].Viewers += tempGame.Viewers
				found = true
			}

		}

		if !found {
			GameList = append(GameList, tempGame)
		}
	}

	return Streams
}

func ShowMainMenu(currGame string, currStreamer string, Streams forms.StreamS, GameList []forms.Games, StreamerList []forms.Streamers) {

	termbox.Clear(coldef, coldef)
	termbox.SetCursor(0, 0)
	termbox.Flush()

	putLn(0, 0, "*** TwitchGo Menu ***", coldef, coldef)
	putLn(0, 1, "Current Game: "+currGame, coldef, coldef)
	putLn(0, 2, "Current Streamer: "+currStreamer, coldef, coldef)

	putLn(0, 4, "1. View Stream", coldef, coldef)
	putLn(0, 5, "2. View Chat", coldef, coldef)
	putLn(0, 6, "4. Change Streamer", coldef, coldef)
	putLn(0, 7, "5. Select Random Stream", coldef, coldef)
	putLn(0, 8, "6. Refresh Streams List", coldef, coldef)

	putLn(0, 10, "x. Quit", coldef, coldef)

	putLn(0, 12, "Select: ", coldef, coldef)

	termbox.SetCursor(8, 12)
mainloop:
	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case termbox.KeyEsc:
				break mainloop
			default:
				if ev.Ch != 0 {
				}
			}
		case termbox.EventError:
			panic(ev.Err)
		}
	}

	// var input string
	// fmt.Scan(&input)

	// if input == "1" {
	// 	fmt.Println("\nOpening stream ...")
	// 	exec.Command("cmd", "/C", "livestreamer http://www.twitch.tv/"+currStreamer+" best").Start()
	// }
	// if input == "2" {
	// 	//do 2 stuff
	// }
	// if input == "3" {
	// 	//do 3 stuff
	// }
	// if input == "4" {

	// 	var intin int
	// 	intin = -1
	// 	for intin < 0 || intin >= len(StreamerList) {
	// 		termbox.Clear(coldef, coldef)
	// 		fmt.Printf("*** Choose Streamer ***\n")
	// 		for i := 0; i < len(StreamerList); i++ {
	// 			fmt.Println(" " + strconv.Itoa(i) + " : " + StreamerList[i].Name + " (" + strconv.Itoa(StreamerList[i].Viewers) + ")")
	// 		}
	// 		fmt.Printf("\nSelect: ")
	// 		fmt.Scan(&intin)
	// 	}

	// 	currStreamer = StreamerList[intin].Name
	// 	currGame = StreamerList[intin].Game
	// }
	// if input == "5" {
	// 	val := rand.Intn(len(Streams.Streams))
	// 	currGame = Streams.Streams[val].Game
	// 	currStreamer = Streams.Streams[val].Channel.Name
	// }
	// if input == "6" {
	// 	//do 6 stuff
	// }
	// if input == "x" {
	// 	Quit = true
	// }
}

func putCh(x int, y int, ch rune, fg termbox.Attribute, bg termbox.Attribute) {
	termbox.SetCell(x, y, ch, coldef, coldef)
	termbox.Flush()
}

func putLn(x int, y int, str string, fg termbox.Attribute, bg termbox.Attribute) {
	cursx = x
	cursy = y
	pos := 0
	for {
		ch, _ := utf8.DecodeRuneInString(str[pos:])
		putCh(cursx, cursy, ch, coldef, coldef)
		cursx += 1
		pos += 1
		if pos >= len(str) {
			break
		}
	}
}

func drawInitStreamList() {
	pos := 0
	cursx := 1
	cursy := 0
	text = "Getting initial stream list..."
	putCh(0, 1, '┌', coldef, coldef)
	putCh(0, 2, '│', coldef, coldef)
	putCh(0, 3, '└', coldef, coldef)

	for {
		ch, _ := utf8.DecodeRuneInString(text[pos:])
		termbox.SetCursor(cursx, cursy)
		putCh(cursx-1, cursy, ch, coldef, coldef)
		time.Sleep(time.Millisecond * 25)

		if pos > 0 {
			putCh(pos, 1, '─', coldef, coldef)
			putCh(pos, 2, ' ', coldef, coldef)
			putCh(pos, 3, '─', coldef, coldef)
		}

		putCh(pos+1, 1, '┐', coldef, coldef)
		putCh(pos+1, 2, '│', coldef, coldef)
		putCh(pos+1, 3, '┘', coldef, coldef)
		cursx += 1
		pos += 1
		if pos >= len(text) {
			break
		}
	}

	pos += 1

	putCh(pos, 1, '╒', coldef, coldef)
	putCh(pos, 2, '│', coldef, coldef)
	putCh(pos, 3, '╘', coldef, coldef)

	pos += 1

	for pos < 36 {
		time.Sleep(time.Millisecond * 25)
		putCh(pos, 1, '═', coldef, coldef)
		putCh(pos, 2, ' ', coldef, coldef)
		putCh(pos, 3, '═', coldef, coldef)
		pos += 1
	}

	putCh(pos, 1, '╕', coldef, coldef)
	putCh(pos-1, 2, '%', coldef, coldef)
	putCh(pos, 2, '│', coldef, coldef)
	putCh(pos, 3, '╛', coldef, coldef)

}

package main

import (
	"encoding/json"
	"fmt"
	"github.com/nsf/termbox-go"
	"github.com/ripture/twitch/lib"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"
	"unicode/utf8"
)

//Global Variables
var text string
var cursx int
var cursy int
var GameList []forms.Games
var StreamerList []forms.Streamers
var currentMenuChoice int

const coldef = termbox.ColorDefault

func main() {

	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()

	termbox.SetInputMode(termbox.InputEsc | termbox.InputMouse)

	Streams := GetStreams(525)

	ShowMainMenu(Streams)
}

func processArgs() {
	args := os.Args[1:]

	if len(args) > 0 {
		if args[0] == "/?" || args[0] == "?" {
			fmt.Printf("\nTwitchGo Usage: twitch -s [streamer name] -g [game name]\n\n")
			return
		}
	}
}

func GetStreams(num int) forms.StreamS {
	var SomeStreams forms.StreamS
	var Streams forms.StreamS
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
			panic(err)
		}

		body, err := ioutil.ReadAll(resp.Body)

		err = json.Unmarshal(body, &SomeStreams)
		if err != nil {
			panic(err)
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
			panic(err)
		}

		body, err := ioutil.ReadAll(resp.Body)

		err = json.Unmarshal(body, &SomeStreams)
		if err != nil {
			panic(err)
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
	return Streams
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

func ShowMainMenu(Streams forms.StreamS) {
	var choice int

	rand.Seed(time.Now().UTC().UnixNano())
	// val := rand.Intn(len(Streams.Streams))
	// currGame := Streams.Streams[val].Game
	// currStreamer := Streams.Streams[val].Channel.Name

	termbox.Clear(coldef, coldef)
	termbox.SetCursor(-1, -1)
	termbox.Flush()

	putLn(14, 0, "╓──╜ TwitchGo Menu ╙──╖", coldef, coldef)
	putLn(14, 1, "╚═════════════════════╝", coldef, coldef)

	putLn(5, 2, "Click or use arrow keys to make selection", termbox.ColorCyan, coldef)

	drawMenuOptionBoxes()

	// putLn(0, 2, "Current Game: "+currGame, coldef, coldef)
	// putLn(0, 3, "Current Streamer: "+currStreamer, coldef, coldef)

	// putLn(0, 5, "1. View Stream", coldef, coldef)
	// putLn(0, 6, "2. View Chat", coldef, coldef)
	// putLn(0, 7, "4. Change Streamer", coldef, coldef)
	// putLn(0, 8, "5. Select Random Stream", coldef, coldef)
	// putLn(0, 9, "6. Refresh Streams List", coldef, coldef)

	// putLn(0, 11, "x. Quit", coldef, coldef)

	// putLn(0, 13, "Select: ", coldef, coldef)

mainloop:
	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case termbox.KeyEsc:
				break mainloop
			case termbox.KeyArrowDown:
				choice = keyChoiceDecode("down")
				drawMenuOptionSelect(choice)
			case termbox.KeyArrowLeft:
				choice = keyChoiceDecode("left")
				drawMenuOptionSelect(choice)
			case termbox.KeyArrowRight:
				choice = keyChoiceDecode("right")
				drawMenuOptionSelect(choice)
			case termbox.KeyArrowUp:
				choice = keyChoiceDecode("up")
				drawMenuOptionSelect(choice)
			}
		case termbox.EventMouse:
			// choice = mouseChoiceDecode(ev.MouseX, ev.MouseY)
			// drawMenuOptionSelect(choice)
			termbox.SetCursor(0, 20)
			fmt.Printf("x-%v, y-%v", ev.MouseX, ev.MouseY)
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
	termbox.SetCell(x, y, ch, fg, bg)
	termbox.Flush()
}

func putLn(x int, y int, str string, fg termbox.Attribute, bg termbox.Attribute) {
	cursx = x
	cursy = y
	var size int

	for {
		ch, size := utf8.DecodeRuneInString(str[size:])
		putCh(cursx, cursy, ch, fg, bg)
		cursx += 1
		str = str[size:]
		if len(str) == 0 {
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

func drawMenuOptionBoxes() {

	posx := 0
	posy := 4

	for i := 0; i < 3; i++ {
		putLn(posx, (3*i)+posy+0, "┌───────────────────────┐", coldef, coldef)
		putLn(posx, (3*i)+posy+1, "│                       │", coldef, coldef)
		putLn(posx, (3*i)+posy+2, "└───────────────────────┘", coldef, coldef)
	}

	for i := 0; i < 3; i++ {
		putLn(posx+25, (3*i)+posy+0, "┌───────────────────────┐", coldef, coldef)
		putLn(posx+25, (3*i)+posy+1, "│                       │", coldef, coldef)
		putLn(posx+25, (3*i)+posy+2, "└───────────────────────┘", coldef, coldef)
	}
}

func mouseChoiceDecode(mouseX int, mouseY int) int {
	//layout is as follows for returns
	//	1 		2
	//	3		4
	//	5		6
	//0 is no option selected
	currentMenuChoice = 0

	if mouseY >= 4 && mouseY <= 6 && mouseX >= 0 && mouseX <= 24 {
		currentMenuChoice = 1
	}
	if mouseY >= 4 && mouseY <= 6 && mouseX >= 25 && mouseX <= 49 {
		currentMenuChoice = 2
	}
	if mouseY >= 7 && mouseY <= 9 && mouseX >= 0 && mouseX <= 24 {
		currentMenuChoice = 3
	}
	if mouseY >= 7 && mouseY <= 9 && mouseX >= 25 && mouseX <= 49 {
		currentMenuChoice = 4
	}
	if mouseY >= 10 && mouseY <= 12 && mouseX >= 0 && mouseX <= 24 {
		currentMenuChoice = 5
	}
	if mouseY >= 10 && mouseY <= 12 && mouseX >= 25 && mouseX <= 49 {
		currentMenuChoice = 6
	}

	return currentMenuChoice
}

func keyChoiceDecode(dir string) int {
	var choice int

	switch dir {
	case "up":
		switch currentMenuChoice {
		case 0:
			choice = 5
		case 1:
			choice = 5
		case 2:
			choice = 6
		case 3:
			choice = 1
		case 4:
			choice = 2
		case 5:
			choice = 3
		case 6:
			choice = 4
		}
	case "down":
		switch currentMenuChoice {
		case 0:
			choice = 1
		case 1:
			choice = 3
		case 2:
			choice = 4
		case 3:
			choice = 5
		case 4:
			choice = 6
		case 5:
			choice = 1
		case 6:
			choice = 2
		}
	case "right":
		switch currentMenuChoice {
		case 0:
			choice = 1
		case 1:
			choice = 2
		case 2:
			choice = 1
		case 3:
			choice = 4
		case 4:
			choice = 3
		case 5:
			choice = 6
		case 6:
			choice = 5
		}
	case "left":
		switch currentMenuChoice {
		case 0:
			choice = 2
		case 1:
			choice = 2
		case 2:
			choice = 1
		case 3:
			choice = 4
		case 4:
			choice = 3
		case 5:
			choice = 6
		case 6:
			choice = 5
		}
	}
	currentMenuChoice = choice
	return choice
}

func drawMenuOptionSelect(choice int) {
	drawMenuOptionBoxes()

	var tbx int
	var tby int
	var bbx int
	var bby int
	var mlx int
	var mly int
	var mrx int
	var mry int

	switch choice {
	case 0:
		return //do no selecting
	case 1:
		tbx = 0
		tby = 4
		bbx = 0
		bby = 6
		mlx = 0
		mly = 5
		mrx = 24
		mry = 5
	case 2:
		tbx = 25
		tby = 4
		bbx = 25
		bby = 6
		mlx = 25
		mly = 5
		mrx = 49
		mry = 5
	case 3:
		tbx = 0
		tby = 7
		bbx = 0
		bby = 9
		mlx = 0
		mly = 8
		mrx = 24
		mry = 8
	case 4:
		tbx = 25
		tby = 7
		bbx = 25
		bby = 9
		mlx = 25
		mly = 8
		mrx = 49
		mry = 8
	case 5:
		tbx = 0
		tby = 10
		bbx = 0
		bby = 12
		mlx = 0
		mly = 11
		mrx = 24
		mry = 11
	case 6:
		tbx = 25
		tby = 10
		bbx = 25
		bby = 12
		mlx = 25
		mly = 11
		mrx = 49
		mry = 11
	}

	putLn(tbx, tby, "┌───────────────────────┐", termbox.ColorGreen, coldef)
	putLn(mlx, mly, "│", termbox.ColorGreen, coldef)
	putLn(mrx, mry, "│", termbox.ColorGreen, coldef)
	putLn(bbx, bby, "└───────────────────────┘", termbox.ColorGreen, coldef)
	putCh(mlx+1, mly, '►', coldef, coldef)
	putCh(mrx-1, mry, '◄', coldef, coldef)
}

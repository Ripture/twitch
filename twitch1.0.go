package main

import (
	"./lib"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"time"
)

func main() {

	args := os.Args[1:]

	if len(args) > 0 {
		if args[0] == "/?" || args[0] == "?" {
			fmt.Printf("\nTwitchGo Usage: twitch -s [streamer name] -g [game name]\n\n")
			return
		}
	}

	ClearScreen()
	var GameList []forms.Games
	var StreamerList []forms.Streamers

	fmt.Println("Getting initial stream list ...")
	Streams, GameList, StreamerList, err := GetStreams(100, GameList, StreamerList)

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

func GetStreams(num int, GameList []forms.Games, StreamerList []forms.Streamers) (forms.StreamS, []forms.Games, []forms.Streamers, error) {
	var SomeStreams forms.StreamS
	var Streams forms.StreamS
	var err error
	var offset int
	var prog float32
	var bars float32
	var progbar string

	bars = 0.0
	prog = 0.0
	fmt.Print(".........................  0%")

	baseURL := "https://api.twitch.tv/kraken"

	numt := num / 100 //number of times we need to get max limit (100) GETs
	numm := num % 100 //number to pick up the remainder of limit for GET (will be limit=numm)

	for i := 0; i < numt; i++ {
		progbar = ""
		offset = 100 * i
		resp, err := http.Get(baseURL + "/streams?limit=100&offset=" + strconv.Itoa(offset))
		defer resp.Body.Close()
		if err != nil {
			fmt.Println(err)
			return Streams, GameList, StreamerList, err
		}

		body, err := ioutil.ReadAll(resp.Body)

		err = json.Unmarshal(body, &SomeStreams)
		if err != nil {
			fmt.Println("error:", err)
			return Streams, GameList, StreamerList, err
		}

		Streams, GameList, StreamerList = BuildStreamList(Streams, SomeStreams, GameList, StreamerList)

		prog += (100 / float32(num)) * 100
		bars = prog / 4.0
		numbars := int(bars)
		for i := numbars; i > 0; i-- {
			progbar += "|"
		}
		numdots := 25 - numbars
		for i := numdots; i > 0; i-- {
			progbar += "."
		}
		progbar = progbar + " " + strconv.Itoa(int(prog))
		fmt.Print("\r" + progbar + "%")
	}

	offset = numt * 100

	if numm > 0 { //if there are any remaining streams to get
		resp, err := http.Get(baseURL + "/streams?limit=" + strconv.Itoa(numm) + "&offset=" + strconv.Itoa(offset))
		if err != nil {
			fmt.Println(err)
			return Streams, GameList, StreamerList, err
		}

		body, err := ioutil.ReadAll(resp.Body)

		err = json.Unmarshal(body, &SomeStreams)
		if err != nil {
			fmt.Println("error:", err)
			return Streams, GameList, StreamerList, err
		}

		Streams, GameList, StreamerList = BuildStreamList(Streams, SomeStreams, GameList, StreamerList)
		fmt.Print("\r||||||||||||||||||||||||| 100%")
	}

	time.Sleep(1 * time.Second)
	return Streams, GameList, StreamerList, err
}

func BuildStreamList(Streams forms.StreamS, SomeStreams forms.StreamS, GameList []forms.Games, StreamerList []forms.Streamers) (forms.StreamS, []forms.Games, []forms.Streamers) {

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

	return Streams, GameList, StreamerList
}

func ShowMainMenu(currGame string, currStreamer string, Streams forms.StreamS, GameList []forms.Games, StreamerList []forms.Streamers) {
	Quit := false

	for !Quit {
		ClearScreen()

		fmt.Printf("*** TwitchGo Menu ***\n")
		fmt.Printf("Current Game: " + currGame + "\n")
		fmt.Printf("Current Streamer: " + currStreamer + "\n\n")
		fmt.Printf("1. View Stream\n")
		fmt.Printf("2. View Chat\n")
		fmt.Printf("3. Change Game\n")
		fmt.Printf("4. Change Streamer\n")
		fmt.Printf("5. Select Random Stream\n")
		fmt.Printf("6. Refresh Streams List\n\n")
		fmt.Printf("x. Quit\n\n")

		fmt.Print("Select: ")
		var input string
		fmt.Scan(&input)

		if input == "1" {
			fmt.Println("\nOpening stream ...")
			exec.Command("cmd", "/C", "livestreamer http://www.twitch.tv/"+currStreamer+" best").Start()
		}
		if input == "2" {
			//do 2 stuff
		}
		if input == "3" {
			//do 3 stuff
		}
		if input == "4" {

			var intin int
			intin = -1
			for intin < 0 || intin >= len(StreamerList) {
				ClearScreen()
				fmt.Printf("*** Choose Streamer ***\n")
				for i := 0; i < len(StreamerList); i++ {
					fmt.Println(" " + strconv.Itoa(i) + " : " + StreamerList[i].Name + " (" + strconv.Itoa(StreamerList[i].Viewers) + ")")
				}
				fmt.Printf("\nSelect: ")
				fmt.Scan(&intin)
			}

			currStreamer = StreamerList[intin].Name
			currGame = StreamerList[intin].Game
		}
		if input == "5" {
			val := rand.Intn(len(Streams.Streams))
			currGame = Streams.Streams[val].Game
			currStreamer = Streams.Streams[val].Channel.Name
		}
		if input == "6" {
			//do 6 stuff
		}
		if input == "x" {
			Quit = true
		}
	}
}

func ClearScreen() {
	cmd := exec.Command("cmd", "/c", "cls")
	cmd.Stdout = os.Stdout
	cmd.Run()
}

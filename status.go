package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	//"net/url"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

const (
	commandCurrentTrack string = "Tell Application \"iTunes\" to (get name of current track) & \"\n\" & (get artist of current track)"
	hipchatRoot         string = "https://api.hipchat.com/v2/"
)

type Track struct {
	Title  string
	Artist string
}

var userID, token string

func main() {
	trackC := make(chan Track)
	tickerC := time.NewTicker(time.Second * 2).C
	var currentTrack Track

	fmt.Println("You can find your personal token on https://hubbiz.hipchat.com/account/api")
	fmt.Print("Token: ")
	fmt.Scan(&token)

	fmt.Print("User Id or Email: ")
	fmt.Scan(&userID)

	for {
		select {
		case newTrack := <-trackC:
			status := "♫  " + newTrack.Title + " - " + newTrack.Artist
			log.Println("Changing status to:", status)
			changeHipchatStatus(status)
		case <-tickerC:
			current, err := exec.Command("/usr/bin/osascript", "-e", commandCurrentTrack).Output()
			if err != nil {
				log.Println("Itunes Stoped or Paused", err)
			} else {
				t := strings.Split(string(current), "\n")
				if currentTrack.Title != t[0] {
					currentTrack = Track{t[0], t[1]}
					go func() {
						trackC <- currentTrack
					}()
				}
			}
		}
	}
}

func clientRequest(method, url string, data string) ([]byte, error) {
	requestURL := hipchatRoot + url + "?auth_token=" + token
	client := &http.Client{}
	request, err := http.NewRequest(method, requestURL, strings.NewReader(data))
	request.Header.Set("Content-Type", "application/json")
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	} else {
		defer response.Body.Close()
		return ioutil.ReadAll(response.Body)
	}
}

func changeHipchatStatus(status string) {
	var data struct {
		Name         string `json:"name"`
		Tile         string `json:"title"`
		MentionName  string `json:"mention_name"`
		IsGroupAdmin bool   `json:"is_group_admin"`
		Timezone     string `json:"timezone"`
		Email        string `json:"email"`
		Presence     struct {
			Status string  `json:"status"`
			Show   *string `json:"show"`
		} `json:"presence"`
	}

	getData, err := clientRequest("GET", "user/"+userID, "")
	if err != nil {
		log.Fatal(err)
	}

	json.Unmarshal(getData, &data)

	data.Presence.Status = status
	params, _ := json.Marshal(&data)

	_, err = clientRequest("PUT", "user/"+userID, string(params))
	if err != nil {
		log.Fatal(err)
	}
}

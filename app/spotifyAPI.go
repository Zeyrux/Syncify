package app

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syncify/utils"
)

type APICredentials struct {
	SpotifyAPIID     string `json:"spotify_api_id"`
	SpotifyAPISecret string `json:"spotify_api_secret"`
}

type AccessToken struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

type SpotifyAPI struct {
	credentials       APICredentials
	accessToken       AccessToken
	headerSearchKey   string
	headerSearchValue string
}

type Track struct {
	Id           string `json:"id"`
	Name         string `json:"name"`
	path         string
	pathImplicit string
}

type Item struct {
	Track Track `json:"track"`
}

type Playlist struct {
	Id           string `json:"id"`
	Name         string `json:"name,omitempty"`
	Items        []Item `json:"items,omitempty"`
	pathPlaylist string
}

// APICredentials
func NewAPICredentials() APICredentials {
	file, err := os.ReadFile("./secrets.json")
	utils.HandleError(err)
	apiCredentials := APICredentials{}
	err = json.Unmarshal(file, &apiCredentials)
	utils.HandleError(err)
	return apiCredentials
}

// SpotifyAPI
func NewSpotifyAPI() *SpotifyAPI {
	var spotifyAPI SpotifyAPI
	spotifyAPI.credentials = NewAPICredentials()
	spotifyAPI.Authenticate()
	return &spotifyAPI
}

func (spotifyAPI *SpotifyAPI) Authenticate() {
	utils.DefaultLog("Authenticating Spotify API")
	base64_credentials := base64.StdEncoding.EncodeToString(
		[]byte(spotifyAPI.credentials.SpotifyAPIID + ":" + spotifyAPI.credentials.SpotifyAPISecret),
	)
	request, err := http.NewRequest(
		"POST",
		"https://accounts.spotify.com/api/token",
		bytes.NewBuffer([]byte("grant_type=client_credentials")),
	)
	utils.HandleError(err)
	request.Header.Add("Authorization", "basic "+base64_credentials)
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	response, err := http.DefaultClient.Do(request)
	utils.HandleError(err)
	responseBody, err := io.ReadAll(response.Body)
	utils.HandleError(err)
	var accessToken AccessToken
	json.Unmarshal(responseBody, &accessToken)
	spotifyAPI.accessToken = accessToken
	spotifyAPI.headerSearchKey = "Authorization"
	spotifyAPI.headerSearchValue = spotifyAPI.accessToken.TokenType + " " + spotifyAPI.accessToken.AccessToken
	utils.DefaultLog("Spotify API authenticated")
}

func (playlist *Playlist) SyncWithSpotify(spotifyAPI *SpotifyAPI) {
	// Get Playlist Data
	utils.DefaultLog("Getting playlist: " + playlist.Id)
	urlSplit := strings.Split(playlist.Id, "/")
	id := urlSplit[len(urlSplit)-1]
	requestData, err := http.NewRequest(
		"GET",
		"https://api.spotify.com/v1/playlists/"+id,
		nil,
	)
	utils.HandleError(err)
	requestData.Header.Add(spotifyAPI.headerSearchKey, spotifyAPI.headerSearchValue)
	responseData, err := http.DefaultClient.Do(requestData)
	utils.HandleError(err)
	responseDataBody, err := io.ReadAll(responseData.Body)
	utils.HandleError(err)
	json.Unmarshal(responseDataBody, &playlist)
	utils.DefaultLog("Playlist gotten: " + playlist.Name)
	// Get Playlist Items
	utils.DefaultLog("Getting playlist items: " + playlist.Name)
	requestItems, err := http.NewRequest(
		"GET",
		"https://api.spotify.com/v1/playlists/"+playlist.Id+"/tracks",
		nil,
	)
	utils.HandleError(err)
	requestItems.Header.Add(spotifyAPI.headerSearchKey, spotifyAPI.headerSearchValue)
	responseItems, err := http.DefaultClient.Do(requestItems)
	utils.HandleError(err)
	responseItemsBody, err := io.ReadAll(responseItems.Body)
	utils.HandleError(err)
	json.Unmarshal(responseItemsBody, &playlist)
	utils.DefaultLog("Playlist items gotten: " + playlist.Name)
}

// Playlist
func (playlist *Playlist) SetPaths(config *Config) {
	playlist.pathPlaylist = config.RootDir + "\\" + playlist.Name
	for i, item := range playlist.Items {
		playlist.Items[i].Track.path = playlist.pathPlaylist + "\\" + item.Track.Name + ".mp3"
		playlist.Items[i].Track.pathImplicit = playlist.pathPlaylist + "\\{title}"
	}
}

// func (playlist *Playlist) Initialize() {
// 	utils.DefaultLog("Initializing playlist: " + playlist.Name)
// 	err := os.MkdirAll(playlist.pathPlaylist, 0755)
// 	utils.HandleError(err)
// 	cmd := exec.Command(
// 		".venv\\Scripts\\spotdl",
// 		"sync",
// 		"https://open.spotify.com/playlist/"+playlist.Id,
// 		"--save-file",
// 		playlist.pathSyncFile,
// 		"--output",
// 		playlist.pathPlaylist+"\\{title}.mp3",
// 	)
// 	err = cmd.Run()
// 	utils.HandleError(err)
// 	playlist.noUpdate = true
// 	utils.DefaultLog("Playlist initialized: " + playlist.Name)
// }

func (playlist *Playlist) Check(config *Config) *Playlist {
	utils.DefaultLog("Checking playlist: " + playlist.Name)
	playlist.SyncWithSpotify(config.spotifyAPI)
	playlist.SetPaths(config)
	if _, err := os.Stat(playlist.pathPlaylist); os.IsNotExist(err) {
		os.MkdirAll(playlist.pathPlaylist, 0755)
	}
	var waitGroup sync.WaitGroup
	for _, item := range playlist.Items {
		config.currentParallelDownlaods <- struct{}{}
		waitGroup.Add(1)
		go item.Track.Check(config, &waitGroup)
	}
	waitGroup.Wait()
	utils.DefaultLog("Playlist checked: " + playlist.Name)
	return playlist
}

func (track *Track) Check(config *Config, waitGroup *sync.WaitGroup) {
	if _, err := os.Stat(track.path); os.IsNotExist(err) {
		utils.DefaultLog("Downloading track: " + track.Name)
		cmd := exec.Command(
			".venv\\Scripts\\spotdl",
			"download",
			"https://open.spotify.com/track/"+track.Id,
			"--output",
			track.pathImplicit,
		)
		err := cmd.Run()
		if err != nil {
			utils.DefaultLog("ERROR downloading track: " + track.Name)
		} else {
			utils.DefaultLog("Track downloaded: " + track.Name)
		}
	}
	<-config.currentParallelDownlaods
	waitGroup.Done()
}

// func (playlist *Playlist) Update(config *Config) {
// 	utils.DefaultLog("Updating playlist: " + playlist.Name)
// 	if !playlist.noUpdate {
// 		fmt.Println(".venv\\Scripts\\spotdl",
// 			"sync",
// 			playlist.pathSyncFile,
// 			"--output",
// 			playlist.pathPlaylist+"\\{title}.mp3")
// 		cmd := exec.Command(
// 			".venv\\Scripts\\spotdl",
// 			"sync",
// 			playlist.pathSyncFile,
// 			"--output",
// 			playlist.pathPlaylist+"\\{title}.mp3",
// 		)
// 		err := cmd.Run()
// 		utils.HandleError(err)
// 	}
// 	utils.DefaultLog("Playlist updated: " + playlist.Name)
// }

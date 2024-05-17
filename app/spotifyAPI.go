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

type Playlist struct {
	Id           string `json:"id"`
	Name         string `json:"name,omitempty"`
	pathPlaylist string
	pathSyncFile string
	noUpdate     bool
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

func (spotifyAPI *SpotifyAPI) GetPlaylist(url string) *Playlist {
	utils.DefaultLog("Getting playlist: " + url)
	urlSplit := strings.Split(url, "/")
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
	var playlist Playlist
	json.Unmarshal(responseDataBody, &playlist)
	utils.DefaultLog("Playlist gotten: " + playlist.Name)
	return &playlist
}

// Playlist
func (playlist *Playlist) SetPaths(config *Config) {
	playlist.pathPlaylist = config.RootDir + "\\" + playlist.Name
	playlist.pathSyncFile = playlist.pathPlaylist + "\\" + ".syncify.spotdl"
}

func (playlist *Playlist) Initialize() {
	utils.DefaultLog("Initializing playlist: " + playlist.Name)
	err := os.MkdirAll(playlist.pathPlaylist, 0755)
	utils.HandleError(err)
	cmd := exec.Command(
		".venv\\Scripts\\spotdl",
		"sync",
		"https://open.spotify.com/playlist/"+playlist.Id,
		"--save-file",
		playlist.pathSyncFile,
		"--output",
		playlist.pathPlaylist+"\\{title}.mp3",
	)
	err = cmd.Run()
	utils.HandleError(err)
	playlist.noUpdate = true
	utils.DefaultLog("Playlist initialized: " + playlist.Name)
}

func (playlist *Playlist) Check(config *Config) *Playlist {
	utils.DefaultLog("Checking playlist: " + playlist.Name)
	if playlist.Name == "" || strings.HasPrefix(playlist.Id, "http") {
		playlist = config.spotifyAPI.GetPlaylist(playlist.Id)
		config.needsSave = true
	}
	playlist.SetPaths(config)
	if _, err := os.Stat(playlist.pathPlaylist); os.IsNotExist(err) {
		playlist.Initialize()
	}
	if _, err := os.Stat(playlist.pathSyncFile); os.IsNotExist(err) {
		playlist.Initialize()
	}
	utils.DefaultLog("Playlist checked: " + playlist.Name)
	return playlist
}

func (playlist *Playlist) Update(config *Config) {
	utils.DefaultLog("Updating playlist: " + playlist.Name)
	if !playlist.noUpdate {
		cmd := exec.Command(
			".venv\\Scripts\\spotdl",
			"sync",
			playlist.pathSyncFile,
			"--output",
			playlist.pathPlaylist+"\\{title}.mp3",
		)
		err := cmd.Run()
		utils.HandleError(err)
	}
	utils.DefaultLog("Playlist updated: " + playlist.Name)
}

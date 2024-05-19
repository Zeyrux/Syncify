package app

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"sync"
	"syncify/utils"
)

const configFile = "./config.json"

type Config struct {
	RootDir                  string     `json:"root_dir"`
	Playlists                []Playlist `json:"playlists"`
	ParallelDownloads        int        `json:"parallel_downloads_per_playlist"`
	spotifyAPI               *SpotifyAPI
	currentParallelDownlaods chan struct{}
}

func NewConfig() *Config {
	utils.DefaultLog("Loading config")
	file, err := os.ReadFile(configFile)
	utils.HandleError(err)
	var config Config
	json.Unmarshal(file, &config)
	utils.DefaultLog("Config loaded")
	config.currentParallelDownlaods = make(chan struct{}, config.ParallelDownloads)
	config.spotifyAPI = NewSpotifyAPI()
	config.spotifyAPI.Authenticate()
	return &config
}

func (config *Config) Save() {
	configCopy := Config{
		RootDir:           config.RootDir,
		Playlists:         config.Playlists,
		ParallelDownloads: config.ParallelDownloads,
		spotifyAPI:        config.spotifyAPI,
	}
	for i := 0; i < len(configCopy.Playlists); i++ {
		configCopy.Playlists[i].Items = nil
	}
	fmt.Println(configCopy.Playlists)
	file, err := json.Marshal(configCopy)
	utils.HandleError(err)
	os.WriteFile(configFile, file, 0644)
}

func (config *Config) Update() {
	var playlistWaitGroup sync.WaitGroup
	for i, playlist := range config.Playlists {
		playlistWaitGroup.Add(1)
		go config.UpdatePlaylist(i, &playlist, &playlistWaitGroup)
	}
	playlistWaitGroup.Wait()
}

func (config *Config) UpdatePlaylist(i int, playlist *Playlist, waitGroup *sync.WaitGroup) {
	config.Playlists[i] = *playlist.Check(config)
	waitGroup.Done()
}

func CheckVenv() {
	log.Default().Println("Checking if venv is installed")
	if _, err := os.Stat(".venv"); os.IsNotExist(err) {
		log.Default().Println("Installing venv...")
		cmd := exec.Command(
			"python",
			"-m",
			"venv",
			".venv",
		)
		err = cmd.Run()
		utils.HandleError(err)
		log.Default().Println("venv installed")
		log.Default().Println("Installing spotdl...")
		cmd = exec.Command(
			".venv\\Scripts\\pip",
			"install",
			"spotdl",
		)
		err = cmd.Run()
		utils.HandleError(err)
		log.Default().Println("spotdl installed")
	} else {
		log.Default().Println("venv is installed")
	}
}

func CheckFFMPEG() {
	log.Default().Println("Checking if ffmpeg is installed")
	homeDirecotry, err := os.UserHomeDir()
	utils.HandleError(err)
	ffmpegPath := homeDirecotry + "/.spotdl/ffmpeg.exe"
	if _, err := os.Stat(ffmpegPath); os.IsNotExist(err) {
		log.Default().Println("Installing ffmpeg...")
		cmd := exec.Command(
			".venv\\Scripts\\spotdl",
			"--download-ffmpeg",
			"--overwrite",
		)
		err = cmd.Run()
		utils.HandleError(err)
		log.Default().Println("ffmpeg installed")
	} else {
		log.Default().Println("ffmpeg is installed")
	}
}

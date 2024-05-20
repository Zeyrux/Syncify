package app

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"sync"
	"syncify/utils"

	"gopkg.in/yaml.v3"
)

// const configFile = "./config.json"
const configFile = "./config.yaml"

type Config struct {
	RootDir                  string     `yaml:"root_dir" json:"root_dir"`
	Playlists                []Playlist `yaml:"playlists" json:"playlists"`
	ParallelDownloads        int        `yaml:"parallel_downloads_per_playlist" json:"parallel_downloads_per_playlist"`
	spotifyAPI               *SpotifyAPI
	currentParallelDownlaods chan struct{}
}

func NewConfig() *Config {
	utils.DefaultLog("Loading config")
	file, err := os.ReadFile(configFile)
	utils.HandleError(err)
	var config Config
	yaml.Unmarshal(file, &config)
	fmt.Println(config)
	utils.DefaultLog("Config loaded")
	config.currentParallelDownlaods = make(chan struct{}, config.ParallelDownloads)
	config.spotifyAPI = NewSpotifyAPI()
	config.spotifyAPI.Authenticate()
	return &config
}

func (config *Config) Save() {
	playlistIds := make([]string, len(config.Playlists))
	for i, playlist := range config.Playlists {
		playlistIds[i] = playlist.Id
	}
	marshaledConfig := utils.MarshalConfigYaml(config.RootDir, config.ParallelDownloads, playlistIds)
	os.WriteFile(configFile, marshaledConfig, 0644)
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

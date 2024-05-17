package app

import (
	"encoding/json"
	"log"
	"os"
	"os/exec"
	"syncify/utils"
)

const configFile = "./config.json"

type Config struct {
	RootDir    string     `json:"root_dir"`
	Playlists  []Playlist `json:"playlists"`
	spotifyAPI *SpotifyAPI
	needsSave  bool
}

func NewConfig() *Config {
	utils.DefaultLog("Loading config")
	file, err := os.ReadFile(configFile)
	utils.HandleError(err)
	var config Config
	json.Unmarshal(file, &config)
	utils.DefaultLog("Config loaded")
	config.spotifyAPI = NewSpotifyAPI()
	config.spotifyAPI.Authenticate()
	return &config
}

func (config Config) Save() {
	file, err := json.Marshal(config)
	utils.HandleError(err)
	os.WriteFile(configFile, file, 0644)
}

func (config Config) Check() {
	for i, playlist := range config.Playlists {
		playlistUpdated := playlist.Check(&config)
		config.Playlists[i] = *playlistUpdated
	}
	if config.needsSave {
		config.Save()
	}
}

func (config Config) Update() {
	for _, playlist := range config.Playlists {
		playlist.Update(&config)
	}
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

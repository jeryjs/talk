package speechUtils

import (
	_ "embed"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	htgotts "github.com/hegedustibor/htgo-tts"
	handlers "github.com/hegedustibor/htgo-tts/handlers"
	"github.com/surfaceyu/edge-tts-go/edgeTTS"
)

//go:embed ..\assets\espeak.exe
var espeakBinary []byte

var espeakBinaryPath string
var ttsSpeech htgotts.Speech

func init() {
	espeakBinaryPath = filepath.Join(os.TempDir(), "espeak.exe")
	if _, err := os.Stat(espeakBinaryPath); os.IsNotExist(err) {
		if err := ioutil.WriteFile(espeakBinaryPath, espeakBinary, 0755); err != nil {
			fmt.Println("Error writing binary file:", err)
			return
		}
	}
}

func SayWithEspeak(text string) {
	exec.Command("taskkill", "/im", "espeak.exe", "/T", "/F").Run()

	cmd := exec.Command(espeakBinaryPath, "--path=Z:/Documents/All-Projects/talk/assets", "-v", "en+f3", "-s", "200", text)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
	}
}

func SayWithTTS(text string) {
	// voices := []string{"en-US-AnaNeural", "zh-CN-YunxiaNeural", "zh-CN-YunxiaNeural", "zh-TW-HsiaoChenNeural", "zh-CN-XiaoyiNeural", "zh-CN-XiaoxiaoNeural"}
	voice := "en-US-AnaNeural"
	tempAudio := os.TempDir()
	tempAudio += "/tempAudio.mp3"
	args := edgeTTS.Args{
		Text:       text,
		Voice:      voice,
		Rate:       "+30%",
		Volume:     "+0%",
		WriteMedia: tempAudio,
	}

	exec.Command("taskkill", "/im", "mpv.com", "/T", "/F").Run()

	edgeTTS.NewTTS(args).AddText(args.Text, args.Voice, args.Rate, args.Volume).Speak()

	// Play temp audio file
	exec.Command("./assets/mpv.com", tempAudio).Run()
	// Remove temp audio file
	os.Remove(tempAudio)
}

func SayWithHtgoTts(text string) {
	ttsSpeech = htgotts.Speech{Folder: "tempAudio", Language: "en-us", Handler: &handlers.Native{}}
	ttsSpeech.Speak(text)
	err := os.RemoveAll("tempAudio")
	if err != nil {
		fmt.Println(err)
	}
}

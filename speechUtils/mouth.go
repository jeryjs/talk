package speechUtils

import (
	_ "embed"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/asticode/go-texttospeech/texttospeech"
	htgotts "github.com/hegedustibor/htgo-tts"
	handlers "github.com/hegedustibor/htgo-tts/handlers"
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
	if err != nil {log.Fatal(err)}
}

func SayWithTTS(text string) {
	texttospeech.NewTextToSpeech().Say(text)
}

func SayWithHtgoTts(text string) {
	ttsSpeech = htgotts.Speech{Folder: "tempAudio", Language: "en", Handler: &handlers.Native{}}
    ttsSpeech.Speak(text)
	err:= os.RemoveAll("tempAudio"); if err != nil {fmt.Println(err)}
}
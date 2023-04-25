package speechUtils

import (
	_ "embed"
	"fmt"
	"io/ioutil"
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

	ttsSpeech = htgotts.Speech{Folder: "tempAudio", Language: "en", Handler: &handlers.Native{}}
}

func SayWithEspeak(text string) {
	exec.Command(espeakBinaryPath, "-v", "en+f3", "-s", "200", text).Start()
}

func SayWithTTS(text string) {
	texttospeech.NewTextToSpeech().Say(text)
}

func SayWithHtgoTts(text string) {
    ttsSpeech.Speak(text)
}
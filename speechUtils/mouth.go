package speechUtils

import (
	_ "embed"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
)

//go:embed ..\assets\espeak.exe
var espeakBinary []byte

var espeakBinaryPath string

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
	cmd := exec.Command(espeakBinaryPath, "-v", "en+f3", "-s", "200", text)
	cmd.Stdout = os.Stdout // set the command's stdout to os.Stdout so that it doesn't interfere with other functions
	cmd.Stderr = os.Stderr // set the command's stderr to os.Stderr so that it doesn't interfere with other functions

	err := cmd.Start() // start the command
	if err != nil {
		fmt.Println("Error starting command:", err)
		return
	}

	err = cmd.Wait() // wait for the command to finish
	if err != nil {
		fmt.Println("Error running command:", err)
		return
	}
}

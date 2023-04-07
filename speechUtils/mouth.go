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

func SayWithEspeak(text string) {
	fmt.Println(text+"\n")

	binaryPath := filepath.Join(os.TempDir(), "espeak.exe")
	if err := ioutil.WriteFile(binaryPath, espeakBinary, 0755); err != nil {
		fmt.Println("Error writing binary file:", err)
		return
	}

	cmd := exec.Command(binaryPath, "-v", "en+f3", "-p", "70", "-s", "200", text)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("Error running command:", err)
		return
	}

	fmt.Println(string(output))
}

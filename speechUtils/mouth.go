package speechUtils

import (
	"fmt"
	"log"
	"os/exec"
)

// Implement Text-To-Speech using eSpeak.exe.
// espeak.exe must exist in same directory as main.go
func Say(text string) {
	fmt.Println(text+"\n")
    var msg = text
    cmd := exec.Command("./espeak.exe", msg)
    if err := cmd.Run(); err != nil {
        log.Fatal(err)
    }
}
package main

import (
	"os"
	"os/exec"
)

func processVideoForFastStart(filepath string) (string, error) {
	outputFile := filepath + ".processing"
	myCmd := exec.Command("ffmpeg", "-i", filepath, "-c", "copy", "-movflags", "faststart", "-f", "mp4", outputFile)

	myCmd.Stderr = os.Stderr

	//var out bytes.Buffer
	//myCmd.Stdout = &out

	err := myCmd.Run()
	if err != nil {
		return "", err
	}

	return outputFile, nil

}

package main

import (
	"bytes"
	"encoding/json"
	"errors"

	//"fmt"
	"os/exec"
)

type Ffprobe struct {
	Streams []FfprobeStream `json:"streams"`
}

type FfprobeStream struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

func getVideoAspectRatio(filepath string) (string, error) {
	var ffprobeResponse Ffprobe

	myCmd := exec.Command("ffprobe", "-v", "error", "-print_format", "json", "-show_streams", filepath)

	var out bytes.Buffer
	myCmd.Stdout = &out

	err := myCmd.Run()
	if err != nil {
		return "", err
	}

	err = json.Unmarshal(out.Bytes(), &ffprobeResponse)
	if err != nil {
		return "", err
	}

	if len(ffprobeResponse.Streams) == 0 {
		return "", errors.New("no streams found in video file")
	}

	w := ffprobeResponse.Streams[0].Width
	h := ffprobeResponse.Streams[0].Height

	//fmt.Println("ffprobe width/height:", w, h)

	if w == h*16/9 {
		return "16:9", nil
	} else if h == w*16/9 {
		return "9:16", nil
	} else {
		return "other", nil
	}

}

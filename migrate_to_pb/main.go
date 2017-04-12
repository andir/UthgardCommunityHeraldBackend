package main

import (
	"fmt"
	"log"
	"os"

	"github.com/andir/UthgardCommunityHeraldBackend/timeseries"

	"path/filepath"
)

func walkHandler(path string, info os.FileInfo, err error) error {
	if info.IsDir() {
		return nil
	}
	log.Println(path)

	timeseries.OpenTimeSeries(path)

	return nil
}

func main() {

	filepath.Walk("./data", walkHandler)

	fmt.Println("vim-go")
}

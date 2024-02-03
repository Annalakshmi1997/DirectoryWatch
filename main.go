package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"gopkg.in/fsnotify.v1"
)

func main() {
	file, err := os.Create("./tempfile.txt")
	fmt.Println(file)
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	writesctring := "This is test file"
	defer watcher.Close()
	_, e := io.WriteString(file, writesctring)
	fmt.Println(e)
	done := make(chan bool)

	go func() {
		for {
			select {
			case event := <-watcher.Events:
				fmt.Println("Event:", event)
				if event.Op&fsnotify.Write == fsnotify.Write {
					fmt.Println("Modified file:", event.Name)
				}
			case err := <-watcher.Errors:
				fmt.Println("Error:", err)
			}
		}
	}()

	err = watcher.Add("C:/DirectoryWatch/tempfile.txt")
	if err != nil {
		log.Fatal(err)
	}

	<-done
}

package main

import (
	"flag"
	"fmt"
	"golang.org/x/exp/inotify"
	"log"
	"os"
)

var usage = `
Usage: %[1]s path [path ...]

Watch the passed paths and log any inotify events to stderr.

Example:
%[1]s /bin
%[1]s /var/log /var/log/*
`

func isDir(path string) bool {
	return false
}

func init() {
	// We want more accurate log times and full date
	log.SetFlags(log.Lmicroseconds | log.Ldate)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, usage, os.Args[0])
	}

	flag.Parse()
	if flag.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "Need to have at least one path to watch")
		flag.Usage()
		os.Exit(1)
	}
}

func main() {
	paths := flag.Args()

	watcher, err := inotify.NewWatcher()
	if err != nil {
		log.Fatal("Unable to create a watcher", err)
	}

	for _, path := range paths {
		err = watcher.Watch(path)
		if err != nil {
			log.Fatalf("Unable to watch '%s': %s", path, err)
		}
	}

	log.Println("Starting monitoring")
	for {
		select {
		case ev := <-watcher.Event:
			log.Printf("Recieved an Event: %s", ev.String())
			/*
				log.Printf("Event name: %s", ev.Name)
				log.Printf("Event Mask: %d", ev.Mask)
				log.Printf("Event Cookie: %d", ev.Cookie)
			*/
		case err := <-watcher.Error:
			log.Printf("Recieved an error: %s", err)
		}
	}
}

package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/cretz/bine/tor"
	"github.com/ipsn/go-libtor"
)

func main() {
	// Start tor with some defaults + elevated verbosity
	fmt.Println("Starting and registering onion service, please wait a bit...")
	t, err := tor.Start(nil, &tor.StartConf{ProcessCreator: libtor.Creator, DebugWriter: os.Stderr})
	if err != nil {
		log.Panicf("Failed to start tor: %v", err)
	}
	defer t.Close()

	// Wait at most a few minutes to publish the service
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	// Create an onion service to listen on any port but show as 80
	onion, err := t.Listen(ctx, &tor.ListenConf{RemotePorts: []int{80}, Version3: true})
	if err != nil {
		log.Panicf("Failed to create onion service: %v", err)
	}
	defer onion.Close()

	fmt.Printf("Please open a Tor capable browser and navigate to http://%v.onion\n", onion.ID)
	onionChan := make(chan error)
	go func(c chan error) {
		fs := http.FileServer(http.Dir("./static"))
		http.Handle("/", fs)
		onionSM := http.NewServeMux()
		onionSM.Handle("/", fs)
		s := &http.Server{
			Handler: onionSM,
		}
		err := s.Serve(onion)
		c <- err
	}(onionChan)

	linkChan := make(chan error)
	go func(c chan error, onionId string) {
		router := gin.Default()
		router.Use(gin.BasicAuth(gin.Accounts{
			"admin": "beourguest",
		}))
		router.LoadHTMLGlob("templates/*")

		router.GET("/", func(c *gin.Context) {
			c.HTML(http.StatusOK, "index.html", gin.H{
				"url": fmt.Sprintf("http://%v.onion", onionId),
			})
		})
		router.Run(":8080")
	}(linkChan, onion.ID)

	onionErr := <-onionChan
	log.Fatalln(onionErr)
	linkErr := <-linkChan
	log.Fatalln(linkErr)
}

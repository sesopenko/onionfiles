package main

import (
	"context"
	"crypto"
	ed255192 "crypto/ed25519"
	"crypto/rand"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/cretz/bine/tor"
	"github.com/ipsn/go-libtor"
)

const (
	privKeyPath = "./keys/onionfiles.pem"
	version3    = true
)

type NotFoundErr struct {
}

func (NotFoundErr) Error() string {
	return "not found"
}

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

	var privKey crypto.PrivateKey
	privKey, err = LoadPrivateKeyFromFile(privKeyPath)
	if err != nil {
		if _, ok := err.(NotFoundErr); ok {
			// Create it
			privKey, err = genAndWritePrivateKey()
			if err != nil {
				log.Panicf("Failed to generate secret key: %v", err)
			}
		} else {
			log.Panicf("Failed to read private key: %v", err)
		}
	}

	// Create an onion service to listen on any port but show as 80
	onion, err := t.Listen(ctx, &tor.ListenConf{
		RemotePorts: []int{80},
		Version3:    version3,
		Key:         privKey,
	})
	if err != nil {
		log.Panicf("Failed to create onion service: %v", err)
	}
	defer onion.Close()

	fmt.Printf("Please open a Tor capable browser and navigate to http://%v.onion\n", onion.ID)
	onionChan := make(chan error)
	go func(c chan error) {
		fs := http.FileServer(http.Dir("./static"))
		serveMux := http.NewServeMux()
		serveMux.Handle("/", fs)
		s := &http.Server{
			Handler: serveMux,
		}
		err := s.Serve(onion)
		c <- err
	}(onionChan)

	onionErr := <-onionChan
	log.Fatalln(onionErr)
}
func genAndWritePrivateKey() (ed255192.PrivateKey, error) {
	if _, privateKey, err := ed255192.GenerateKey(rand.Reader); err != nil {
		return nil, err
	} else {
		// Save it
		privatePem, err := os.Create(privKeyPath)
		if err != nil {
			return nil, err
		}
		defer privatePem.Close()
		if _, err = privatePem.Write(privateKey); err != nil {
			return nil, err
		}
		fmt.Println("Generated new private key using ed25519 algorithm")
		return privateKey, nil
	}
}

func LoadPrivateKeyFromFile(filename string) (ed255192.PrivateKey, error) {
	if _, err := os.Stat(filename); err != nil {
		// File doesn't, throw an error
		return nil, NotFoundErr{}
	}
	if privateKeyFile, err := os.Open(filename); err != nil {
		return nil, err
	} else {
		defer privateKeyFile.Close()
		key, err := ioutil.ReadAll(privateKeyFile)
		if err != nil {
			return nil, err
		}
		return key, nil
	}
}

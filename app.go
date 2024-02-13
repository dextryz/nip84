package highlighter

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"os"
	"sync"

	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip19"
)

var ErrNotFound = errors.New("todo list not found")

var KindHighlight = 9802

type Config struct {
	Nsec   string   `json:"nsec"`
	Relays []string `json:"relays"`
}

func LoadConfig(path string) (*Config, error) {

	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("Config file: %v", err)
	}

	var cfg Config
	err = json.Unmarshal(data, &cfg)
	if err != nil {
		log.Fatal(err)
	}

	return &cfg, nil
}

func Publish(ctx context.Context, cfg *Config, content, context, url string) error {

	var sk string
	var pub string
	if _, s, err := nip19.Decode(cfg.Nsec); err == nil {
		sk = s.(string)
		if pub, err = nostr.GetPublicKey(s.(string)); err != nil {
			return err
		}
	} else {
		return err
	}

	e := nostr.Event{
		Kind:      nostr.KindTextNote,
		PubKey:    pub,
		Content:   content,
		CreatedAt: nostr.Now(),
		Tags: nostr.Tags{
			{"r", url},
			//{"context", context},
		},
	}
    err := e.Sign(sk)
	if err != nil {
		return err
	}

    log.Println(e)

	var wg sync.WaitGroup
	for _, r := range cfg.Relays {
		wg.Add(1)

		go func(url string) {
			defer wg.Done()

			relay, err := nostr.RelayConnect(ctx, url)
			if err != nil {
				log.Println(err)
				return
			}
			defer relay.Close()

			err = relay.Publish(ctx, e)
			if err != nil {
				log.Println(err)
				return
			}
		}(r)
	}
	wg.Wait()

	log.Printf("Job request sent to Shipyard DVM")

	return nil
}

func Main() error {

	path, ok := os.LookupEnv("NOSTR")
	if !ok {
		log.Fatalln("NOSTR env var not set")
	}

	cfg, err := LoadConfig(path)
	if err != nil {
		return err
	}

	args := os.Args[1:]

	ctx := context.Background()

    content := args[0]
    context := args[1]
    url := args[2]

	err = Publish(ctx, cfg, content, context, url)
	if err != nil {
		return err
	}

	return nil
}

package highlighter

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
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

type Highlight struct {
	Content string `json:"content"`
	Context string `json:"context"`
	Url     string `json:"url"`
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

func Publish(ctx context.Context, cfg *Config, h Highlight) error {

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
		Kind:      KindHighlight,
		PubKey:    pub,
		Content:   h.Content,
		CreatedAt: nostr.Now(),
		Tags: nostr.Tags{
			{"r", h.Url},
			{"context", h.Context},
		},
	}
	err := e.Sign(sk)
	if err != nil {
		return err
	}

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

	log.Printf("Highlighted event published to nostr relays")

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

	h := Highlight{}

	flag.StringVar(&h.Content, "content", "", "event text note of Kind 1")
	flag.StringVar(&h.Context, "context", "", "event text note of Kind 1")
	flag.StringVar(&h.Url, "url", "", "event text note of Kind 1")

	flag.Parse()
	log.SetFlags(0)

	ctx := context.Background()

	err = Publish(ctx, cfg, h)
	if err != nil {
		return err
	}

	return nil
}

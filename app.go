package highlighter

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"sync"

	nos "github.com/dextryz/nostr"
	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip19"
)

var ErrNotFound = errors.New("todo list not found")

type Highlight struct {
	Content  string `json:"content"`
	Context  string `json:"context"`
	Url      string `json:"url"`
	TextNote string `json:"textnote"`
	Article  string `json:"article"`
}
func (s *Article) ReqHighlights(cfg *nos.Config, naddr string) (*nostr.Event, error) {
}


func Publish(ctx context.Context, cfg *nos.Config, h Highlight) error {

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

	tags := nostr.Tags{
		{"context", h.Context},
	}

	// TODO: Make sure it is a correct URL
	if h.Url != "" {
		tags = append(tags, nostr.Tag{"r", h.Url})
	}

	// TODO: Process naddr, note, nevent strings
	if h.TextNote != "" {
		tags = append(tags, nostr.Tag{"e", h.TextNote})
	}

	// TODO: Process naddr, note, nevent strings
	if h.Article != "" {

		prefix, data, err := nip19.Decode(h.Article)
		if err != nil {
			return err
		}
		if prefix != "naddr" {
			return err
		}
		ep := data.(nostr.EntityPointer)

		if ep.Kind != nostr.KindArticle {
			return err
		}

		log.Println(ep)

        v := fmt.Sprintf("%d:%s:%s", ep.Kind, ep.PublicKey, ep.Identifier)
		tags = append(tags, nostr.Tag{"a", v})
	}

	e := nostr.Event{
		Kind:      nos.KindHighlight,
		PubKey:    pub,
		Content:   h.Content,
		CreatedAt: nostr.Now(),
		Tags:      tags,
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

	log.Printf("Highlight published to nostr relays")

	return nil
}

func Main() error {

	path, ok := os.LookupEnv("NOSTR")
	if !ok {
		log.Fatalln("NOSTR env var not set")
	}

	cfg, err := nos.LoadConfig(path)
	if err != nil {
		return err
	}

	h := Highlight{}

	flag.StringVar(&h.Content, "content", "", "event text note of Kind 1")
	flag.StringVar(&h.Context, "context", "", "event text note of Kind 1")
	flag.StringVar(&h.Url, "url", "", "event text note of Kind 1")
	flag.StringVar(&h.TextNote, "textnote", "", "event text note of Kind 1")
	flag.StringVar(&h.Article, "article", "", "event text note of Kind 1")

	flag.Parse()
	log.SetFlags(0)

	ctx := context.Background()

	err = Publish(ctx, cfg, h)
	if err != nil {
		return err
	}

	return nil
}

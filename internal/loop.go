package internal

import (
	"context"
	"log"

	"github.com/caedis/noreza/internal/input"
	"github.com/caedis/noreza/internal/mapping"
	"github.com/caedis/noreza/internal/output"
)

func RunEventLoop(ctx context.Context, reader *input.Reader, store *mapping.Store, writer *output.Writer) {
	events := make(chan mapping.JoystickEvent, 128)
	go reader.Stream(events)

	for {
		select {
		case <-ctx.Done():
			return
		case evt := <-events:
			if !evt.Ready {
				log.Fatal("read error")
				return
			}

			store.BroadcastEvent(mapping.SSEEvent{Type: mapping.EventJoystick, Data: evt})
			press, release := store.Resolve(evt)
			writer.Apply(press, release)
		}
	}
}

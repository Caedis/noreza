package main

import (
	"context"
	"flag"
	"io"
	"log"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"sync"
	"syscall"
	"time"

	"github.com/caedis/noreza/internal"
	"github.com/caedis/noreza/internal/input"
	"github.com/caedis/noreza/internal/mapping"
	"github.com/caedis/noreza/internal/output"
	"github.com/caedis/noreza/internal/shared/paths"
	"github.com/caedis/noreza/internal/web"
)

var inputSerial = flag.String("serial", "", "serial of target azeron device")
var port = flag.Int("port", 1337, "web server port")
var quiet = flag.Bool("quiet", false, "disable logging")
var wait = flag.Bool("wait", false, "")
var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")
var memprofile = flag.String("memprofile", "", "write memory profile to `file`")

func main() {
	flag.Parse()

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		defer f.Close()
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
	}

	if *quiet {
		log.SetOutput(io.Discard)
	}

	if *inputSerial == "" {
		log.Fatal("No input device serial provided")
	}

	log.Println("Connecting to device")
	var devicePath string
	var productID uint16
	var err error
	var wroteMessage bool
	for {
		devicePath, productID, err = input.GetDevicePath(*inputSerial)
		if err != nil {
			if !*wait {
				log.Fatal(err.Error())
				return
			}
			if !wroteMessage {
				log.Println("Retrying every 2s for device to be connected")
				wroteMessage = true
			}
			time.Sleep(2 * time.Second)
			continue
		}
		break
	}
	log.Print("Connected\n")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	writer, err := output.NewWriter(*inputSerial)
	if err != nil {
		log.Fatalf("error creating writer: %v", err)
	}

	profilesPath := paths.ProfilesDir(*inputSerial)
	if err := os.MkdirAll(profilesPath, 0755); err != nil {
		log.Fatalf("error creating profile directory: %v", err)
	}

	store := mapping.NewStore(profilesPath, productID)
	if err := store.CreateIfNeeded(); err != nil {
		log.Fatal(err)
	}
	if err := store.ReloadAllProfiles(); err != nil {
		log.Fatal("failed to load mapping:", err)
	}
	if err := store.ReloadActive(); err != nil {
		log.Fatal("failed to load active:", err)
	}

	go store.WatchProfiles(ctx)

	reader, err := input.NewReader(devicePath)
	if err != nil {
		log.Fatalf("failed to start reader: %v", err)
	}

	if _, found := os.LookupEnv("WAYLAND_DISPLAY"); found {
		log.Println("Active window watching disabled on wayland")
	} else {
		log.Println("Watching active windows")
		switcher, err := mapping.NewAutoProfileSwitcher(store, 300*time.Millisecond)
		if err != nil {
			log.Fatal(err)
		}
		go switcher.Start(ctx)
	}
	go web.RunServer(ctx, *port, store, reader.String(), *inputSerial)
	go internal.RunEventLoop(ctx, reader, store, writer)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGHUP,
	)

	var mu sync.Mutex
	go func() {
		sig := <-sigs
		log.Printf("[signal] caught %s, shutting down...", sig)
		cancel()

		mu.Lock()
		defer mu.Unlock()

		if reader != nil {
			reader.Close()
		}

		if writer != nil {
			writer.Close()
		}

		os.Exit(0)
	}()

	log.Println("[daemon] started. press Ctrl+C to stop.")
	<-ctx.Done()

	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			log.Fatal("could not create memory profile: ", err)
		}
		defer f.Close()
		runtime.GC() // get up-to-date statistics
		if err := pprof.Lookup("allocs").WriteTo(f, 0); err != nil {
			log.Fatal("could not write memory profile: ", err)
		}
	}
}

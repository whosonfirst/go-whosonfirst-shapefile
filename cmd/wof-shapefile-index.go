package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/feature"
	"github.com/whosonfirst/go-whosonfirst-index"
	"github.com/whosonfirst/go-whosonfirst-index/utils"
	"github.com/whosonfirst/go-whosonfirst-log"
	"github.com/whosonfirst/go-whosonfirst-shapefile"
	"github.com/whosonfirst/warning"
	"io"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

func main() {

	valid_modes := strings.Join(index.Modes(), ",")
	desc_modes := fmt.Sprintf("The mode to use importing data. Valid modes are: %s.", valid_modes)

	mode := flag.String("mode", "repo", desc_modes)
	out := flag.String("out", "", "...")

	timings := flag.Bool("timings", false, "Display timings during and after indexing")

	flag.Parse()

	logger := log.SimpleWOFLogger()

	stdout := io.Writer(os.Stdout)
	logger.AddLogger(stdout, "status")

	writer, err := shapefile.NewPointWriter(*out)

	if err != nil {
		logger.Fatal("Failed to create new shape because %s", err)
	}

	writer.Logger = logger

	/* please move all of this in to a package */

	mu := new(sync.Mutex)

	cb := func(fh io.Reader, ctx context.Context, args ...interface{}) error {

		path, err := index.PathForContext(ctx)

		if err != nil {
			return err
		}

		ok, err := utils.IsPrincipalWOFRecord(fh, ctx)

		if err != nil {
			return err
		}

		if !ok {
			return nil
		}

		f, err := feature.LoadWOFFeatureFromReader(fh)

		if err != nil {

			if err != nil && !warning.IsWarning(err) {
				msg := fmt.Sprintf("Unable to load %s, because %s", path, err)
				return errors.New(msg)
			}
		}

		mu.Lock()
		defer mu.Unlock()

		_, err = writer.AddFeature(f)

		if err != nil {
			return err
		}

		return nil
	}

	indexer, err := index.NewIndexer(*mode, cb)

	if err != nil {
		logger.Fatal("Failed to create new indexer because: %s", err)
	}

	done_ch := make(chan bool)
	t1 := time.Now()

	show_timings := func() {

		t2 := time.Since(t1)

		i := atomic.LoadInt64(&indexer.Indexed) // please just make this part of go-whosonfirst-index
		logger.Status("time to index all (%d) : %v", i, t2)
	}

	if *timings {

		go func() {

			for {

				select {
				case <-done_ch:
					return
				case <-time.After(1 * time.Minute):
					show_timings()
				}
			}
		}()

		defer func() {
			done_ch <- true
		}()
	}

	err = indexer.IndexPaths(flag.Args())

	/* end of please move this in to a proper library */

	if err != nil {
		logger.Fatal("Failed to index paths in %s mode because: %s", *mode, err)
	}

	writer.Close()
	os.Exit(0)
}

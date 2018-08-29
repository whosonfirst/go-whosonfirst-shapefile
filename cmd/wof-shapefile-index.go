package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/whosonfirst/go-whosonfirst-cli/flags"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/feature"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/properties/whosonfirst"
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

	valid_types := strings.Join(shapefile.ShapeTypes(), ",")
	desc_types := fmt.Sprintf("The shapefile type to use indexing data. Valid types are: %s.", valid_types)

	// something something something take the filter code in go-whosonfirst-pip-v2
	// and make it generally applicable to something like this - ultimately it should
	// be made to work with go-whosonfirst-index callback function
	// filters, err := filter.NewSPRFilterFromQuery(query)
	// (20180817/thisisaaronland)

	var include_placetype flags.MultiString
	flag.Var(&include_placetype, "include-placetype", "Include only records of this placetype. You may pass multiple -include-placetype flags.")

	var exclude_placetype flags.MultiString
	flag.Var(&exclude_placetype, "exclude-placetype", "Exclude records of this placetype. You may pass multiple -exclude-placetype flags.")

	var belongs_to flags.MultiInt64
	flag.Var(&belongs_to, "belongs-to", "Include only records that belong to this ID. You may pass multiple -belongs-to flags.")

	mode := flag.String("mode", "repo", desc_modes)

	shapetype := flag.String("shapetype", "POINT", desc_types)

	out := flag.String("out", "", "Where to write the new shapefile")

	timings := flag.Bool("timings", false, "Display timings during and after indexing")

	flag.Parse()

	logger := log.SimpleWOFLogger()

	stdout := io.Writer(os.Stdout)
	logger.AddLogger(stdout, "status")

	writer, err := shapefile.NewWriterFromString(*out, *shapetype)

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

		pt := f.Placetype()

		if len(include_placetype) > 0 {

			if !include_placetype.Contains(pt) {
				return nil
			}
		}

		if len(exclude_placetype) > 0 {

			if exclude_placetype.Contains(pt) {
				return nil
			}
		}

		if len(belongs_to) > 0 {

			ok := false

			for _, candidate := range belongs_to {

				if whosonfirst.IsBelongsTo(f, candidate) {
					ok = true
					break
				}
			}

			if !ok {
				return nil
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

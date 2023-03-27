package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/whosonfirst/go-whosonfirst-iterate/v2/iterator"
	"github.com/whosonfirst/go-whosonfirst-shapefile"
	"github.com/whosonfirst/go-whosonfirst-uri"
)

func main() {

	valid_types := strings.Join(shapefile.ShapeTypes(), ",")
	desc_types := fmt.Sprintf("The shapefile type to use indexing data. Valid types are: %s.", valid_types)

	shapetype := flag.String("shapetype", "POINT", desc_types)

	out := flag.String("out", "", "Where to write the new shapefile")

	iterator_uri := flag.String("iterator-uri", "repo://", "...")
	flag.Parse()

	ctx := context.Background()
	logger := log.Default()

	uris := flag.Args()

	writer, err := shapefile.NewWriterFromString(*out, *shapetype)

	if err != nil {
		logger.Fatalf("Failed to create new shape because %s", err)
	}

	writer.Logger = logger

	/* please move all of this in to a package */

	mu := new(sync.Mutex)

	iter_cb := func(ctx context.Context, path string, r io.ReadSeeker, args ...interface{}) error {

		_, uri_args, err := uri.ParseURI(path)

		if err != nil {
			return fmt.Errorf("Failed to parse %s, %w", path, err)
		}

		if uri_args.IsAlternate {
			return nil
		}

		body, err := io.ReadAll(r)

		if err != nil {
			return fmt.Errorf("Failed to read %s, %w", path, err)
		}

		mu.Lock()
		defer mu.Unlock()

		_, err = writer.AddFeature(body)

		if err != nil {
			return fmt.Errorf("Failed to add %s to shapefile, %w", path, err)
		}

		return nil
	}

	iter, err := iterator.NewIterator(ctx, *iterator_uri, iter_cb)

	if err != nil {
		logger.Fatalf("Failed to create new indexer because: %v", err)
	}

	err = iter.IterateURIs(ctx, uris...)

	if err != nil {
		logger.Fatalf("Failed to iterate URIs, %v", err)
	}

	writer.Close()
	os.Exit(0)
}

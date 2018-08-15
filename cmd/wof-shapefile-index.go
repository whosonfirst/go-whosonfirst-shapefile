package main

import (
	"flag"
	"fmt"
	wof_index "github.com/whosonfirst/go-whosonfirst-index"
	"github.com/whosonfirst/go-whosonfirst-log"
	"github.com/whosonfirst/go-whosonfirst-shapefile"
	"github.com/whosonfirst/go-whosonfirst-shapefile/index"
	"io"
	"os"
	"strings"
)

func main() {

	valid_modes := strings.Join(wof_index.Modes(), ",")
	desc_modes := fmt.Sprintf("The mode to use importing data. Valid modes are: %s.", valid_modes)

	valid_types := strings.Join(shapefile.ShapeTypes(), ",")
	desc_types := fmt.Sprintf("The shapefile type to use indexing data. Valid types are: %s.", valid_types)

	mode := flag.String("mode", "repo", desc_modes)

	shapetype := flag.String("shapetype", "POINT", desc_types)

	out := flag.String("out", "", "Where to write the new shapefile")

	// timings := flag.Bool("timings", false, "Display timings during and after indexing")

	flag.Parse()

	logger := log.SimpleWOFLogger()

	stdout := io.Writer(os.Stdout)
	logger.AddLogger(stdout, "status")

	writer, err := shapefile.NewWriterFromString(*out, *shapetype)

	if err != nil {
		logger.Fatal("Failed to create new shape because %s", err)
	}

	writer.Logger = logger

	idx, err := index.NewShapefileIndexer(*mode, writer)

	if err != nil {
		logger.Fatal(error)
	}

	err = idx.IndexPaths(flag.Args())

	if err != nil {
		logger.Fatal("Failed to index paths in %s mode because: %s", *mode, err)
	}

	writer.Close()
	os.Exit(0)
}

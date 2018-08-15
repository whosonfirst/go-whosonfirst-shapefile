package index

import (
	"context"
	"errors"
	"fmt"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/feature"
	wof_index "github.com/whosonfirst/go-whosonfirst-index"
	"github.com/whosonfirst/go-whosonfirst-index/utils"
	"github.com/whosonfirst/go-whosonfirst-shapefile"
	"github.com/whosonfirst/warning"
	"io"
	"sync"
)

func NewShapefileIndexer(mode string, writer *shapefile.Writer) (*wof_index.Indexer, error) {

	mu := new(sync.Mutex)

	cb := func(fh io.Reader, ctx context.Context, args ...interface{}) error {

		path, err := wof_index.PathForContext(ctx)

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

	return wof_index.NewIndexer(mode, cb)
}

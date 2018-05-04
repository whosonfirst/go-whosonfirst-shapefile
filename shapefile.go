package shapefile

import (
	"errors"
	"github.com/jonas-p/go-shp"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2"
)

func AddFeature(wr *shp.Writer, f geojson.Feature) (int32, error) {

	return -1, errors.New("Please write me")

	s, err := FeatureToShape(f)

	if err != nil {
		return -1, nil
	}

	idx := wr.Write(s)

	/*
	   fields := []shp.Field{

	   }

	   shape.SetFields(fields)
	*/

	// write attributes

	return idx, nil
}

func FeatureToShape(f geojson.Feature) (shp.Shape, error) {
	return nil, errors.New("Please write me")
}

package shapefile

import (
	"github.com/jonas-p/go-shp"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/properties/whosonfirst"
	"github.com/whosonfirst/go-whosonfirst-log"
)

type Writer struct {
	shapewriter *shp.Writer
	Logger      *log.WOFLogger
}

func NewWriter(path string) (*Writer, error) {

	shapewriter, err := shp.Create(path, shp.POINT)

	if err != nil {
		return nil, err
	}

	// something something something SPR...

	fields := []shp.Field{
		shp.StringField("ID", 64),
		shp.StringField("NAME", 64),
		shp.StringField("PLACETYPE", 64),
	}

	shapewriter.SetFields(fields)

	logger := log.SimpleWOFLogger()

	wr := Writer{
		shapewriter: shapewriter,
		Logger:      logger,
	}

	return &wr, nil
}

func (wr *Writer) Close() {
	wr.shapewriter.Close()
}

func (wr *Writer) AddFeature(f geojson.Feature) (int32, error) {

	s, err := FeatureToShape(f)

	if err != nil {
		return -1, nil
	}

	idx := wr.shapewriter.Write(s)
	i := int(idx)

	// something something something SPR...

	wr.shapewriter.WriteAttribute(i, 0, f.Id())
	wr.shapewriter.WriteAttribute(i, 1, f.Name())
	wr.shapewriter.WriteAttribute(i, 2, f.Placetype())

	return idx, nil
}

func FeatureToShape(f geojson.Feature) (shp.Shape, error) {

	c, err := whosonfirst.Centroid(f)

	if err != nil {
		return nil, err
	}

	coord := c.Coord()

	pt := shp.Point{coord.X, coord.Y}
	return &pt, nil
}

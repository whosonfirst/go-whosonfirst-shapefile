package shapefile

// https://www.esri.com/library/whitepapers/pdfs/shapefile.pdf

import (
	"errors"
	"github.com/jonas-p/go-shp"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/properties/whosonfirst"
	"github.com/whosonfirst/go-whosonfirst-log"
	// golog "log"
	"strings"
)

type Writer struct {
	shapewriter *shp.Writer
	shapetype   shp.ShapeType // https://godoc.org/github.com/jonas-p/go-shp#ShapeType
	Logger      *log.WOFLogger
}

func ShapeTypes() []string {

	return []string{
		"POINT",
		"POLYGON",
	}
}

func IsValidShapeType(test string) bool {

	valid := false

	for _, shapetype := range ShapeTypes() {

		if shapetype == test {
			valid = true
			break
		}
	}

	return valid
}

func NewWriterFromString(path string, shapetype string) (*Writer, error) {

	switch strings.ToUpper(shapetype) {

	case "POINT":
		return NewWriter(path, shp.POINT)
	case "POLYGON":
		return NewWriter(path, shp.POLYGON)
	default:
		return nil, errors.New("Unsupported shape type")
	}
}

func NewWriter(path string, shapetype shp.ShapeType) (*Writer, error) {

	shapewriter, err := shp.Create(path, shapetype)

	if err != nil {
		return nil, err
	}

	// something something something SPR...
	// https://github.com/whosonfirst/go-whosonfirst-spr

	// there is also a whole stack of pre-existing logic here that I
	// am not keen to formalize but that should be possible to support:
	// https://github.com/whosonfirst/py-mapzen-whosonfirst-utils/blob/master/scripts/wof-csv-to-feature-collection

	// so maybe some sort of generic AttributesFunction for which we
	// can define a sensible default (SPR, etc) but that people can
	// override to suit their needs... or something like that
	// (20180815/thisisaaronland)
	
	fields := []shp.Field{
		shp.StringField("ID", 64),
		shp.StringField("NAME", 64),
		shp.StringField("PLACETYPE", 64),
		shp.StringField("INCEPTION", 64),
		shp.StringField("CESSATION", 64),
	}

	shapewriter.SetFields(fields)

	logger := log.SimpleWOFLogger()

	wr := Writer{
		shapewriter: shapewriter,
		shapetype:   shapetype,
		Logger:      logger,
	}

	return &wr, nil
}

func (wr *Writer) Close() {
	wr.shapewriter.Close()
}

func (wr *Writer) AddFeature(f geojson.Feature) (int32, error) {

	s, err := FeatureToShape(f, wr.shapetype)

	if err != nil {
		return -1, nil
	}

	idx := wr.shapewriter.Write(s)
	i := int(idx)

	// something something something SPR...
	// see notes about attributes above...
	// (20180815/thisisaaronland)
	
	wr.shapewriter.WriteAttribute(i, 0, f.Id())
	wr.shapewriter.WriteAttribute(i, 1, f.Name())
	wr.shapewriter.WriteAttribute(i, 2, f.Placetype())
	wr.shapewriter.WriteAttribute(i, 3, whosonfirst.Inception(f))
	wr.shapewriter.WriteAttribute(i, 4, whosonfirst.Cessation(f))

	return idx, nil
}

func FeatureToShape(f geojson.Feature, shapetype shp.ShapeType) (shp.Shape, error) {

	switch shapetype {

	case shp.POINT:
		return FeatureToPoint(f)
	case shp.POLYGON:
		return FeatureToPolygon(f)
	default:
		return nil, errors.New("Unsupported shape type")
	}
}

func FeatureToPoint(f geojson.Feature) (shp.Shape, error) {

	c, err := whosonfirst.Centroid(f)

	if err != nil {
		return nil, err
	}

	coord := c.Coord()

	pt := shp.Point{coord.X, coord.Y}
	return &pt, nil
}

func FeatureToPolygon(f geojson.Feature) (shp.Shape, error) {

	polys, err := f.Polygons()

	if err != nil {
		return nil, err
	}

	points := make([][]shp.Point, 0)

	for _, poly := range polys {

		if len(poly.InteriorRings()) > 0 {
			return nil, errors.New("Polygon has interior rings")
		}

		ext := poly.ExteriorRing()

		pts := make([]shp.Point, 0)

		for _, coord := range ext.Vertices() {
			pt := shp.Point{coord.X, coord.Y}
			pts = append(pts, pt)
		}

		points = append(points, pts)
	}

	polygon := shp.NewPolyLine(points)
	return polygon, nil
}

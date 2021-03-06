package shapefile

// https://www.esri.com/library/whitepapers/pdfs/shapefile.pdf

import (
	"errors"
	"github.com/jonas-p/go-shp"
	"github.com/tidwall/gjson"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/properties/whosonfirst"
	"github.com/whosonfirst/go-whosonfirst-log"
	"os"
	"path/filepath"
	"strings"
)

type Writer struct {
	shapewriter *shp.Writer
	shapetype   shp.ShapeType // https://godoc.org/github.com/jonas-p/go-shp#ShapeType
	path        string
	Logger      *log.WOFLogger
}

func ShapeTypes() []string {

	return []string{
		"MULTIPOINT",
		"POINT",
		"POLYGON",
		"POLYLINE",
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

	// https://godoc.org/github.com/jonas-p/go-shp#ShapeType

	switch strings.ToUpper(shapetype) {

	case "MULTIPOINT":
		return NewWriter(path, shp.MULTIPOINT)
	case "POLYLINE":
		return NewWriter(path, shp.POLYLINE)
	case "POINT":
		return NewWriter(path, shp.POINT)
	case "POLYGON":
		return NewWriter(path, shp.POLYGON)
	default:
		return nil, errors.New("Unsupported shape type")
	}
}

func NewWriter(path string, shapetype shp.ShapeType) (*Writer, error) {

	abs_path, err := filepath.Abs(path)

	if err != nil {
		return nil, err
	}

	shapewriter, err := shp.Create(abs_path, shapetype)

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
		path:        abs_path,
	}

	return &wr, nil
}

func (wr *Writer) Close() error {
	wr.shapewriter.Close()
	return wr.WriteProjFile()
}

func (wr *Writer) WriteProjFile() error {

	prj_path := strings.Replace(wr.path, ".shp", ".prj", -1)

	fh, err := os.OpenFile(prj_path, os.O_RDWR|os.O_CREATE, 0644)

	if err != nil {
		return nil
	}

	_, err = fh.Write([]byte(`GEOGCS["GCS_WGS_1984",DATUM["D_WGS_1984",SPHEROID["WGS_1984",6378137,298.257223563]],PRIMEM["Greenwich",0],UNIT["Degree",0.017453292519943295]]`))

	if err != nil {
		return nil
	}

	return fh.Close()
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

	case shp.MULTIPOINT:
		return FeatureToMultiPoint(f)
	case shp.POLYLINE:
		return FeatureToPolyline(f)
	case shp.POINT:
		return FeatureToPoint(f)
	case shp.POLYGON:
		return FeatureToPolygon(f)
	default:
		return nil, errors.New("Unsupported shape type")
	}
}

func FeatureToMultiPoint(f geojson.Feature) (shp.Shape, error) {

	coords := gjson.GetBytes(f.Bytes(), "geometry.coordinates")

	if !coords.Exists() {
		return nil, errors.New("Missing coordinates")
	}

	points := make([]shp.Point, 0)

	swlat := 0.0
	swlon := 0.0
	nelat := 0.0
	nelon := 0.0

	for _, c := range coords.Array() {

		pt := c.Array()

		lat := pt[1].Float()
		lon := pt[0].Float()

		shp_pt := shp.Point{lon, lat}
		points = append(points, shp_pt)

		if lat < swlat {
			swlat = lat
		}

		if lon < swlon {
			swlon = lon
		}

		if lat > nelat {
			lat = nelat
		}

		if lon > nelon {
			lon = nelon
		}

	}

	box := shp.Box{
		swlon, swlat, nelon, nelat,
	}

	num := int32(len(points))

	multi := shp.MultiPoint{
		NumPoints: num,
		Points:    points,
		Box:       box,
	}

	return &multi, nil
}

func FeatureToPolyline(f geojson.Feature) (shp.Shape, error) {

	coords := gjson.GetBytes(f.Bytes(), "geometry.coordinates")

	if !coords.Exists() {
		return nil, errors.New("Missing coordinates")
	}

	points := make([]shp.Point, 0)

	for _, c := range coords.Array() {

		pt := c.Array()

		lat := pt[1].Float()
		lon := pt[0].Float()

		shp_pt := shp.Point{lon, lat}
		points = append(points, shp_pt)
	}

	poly := [][]shp.Point{points}

	polygon := shp.NewPolyLine(poly)
	return polygon, nil
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

		/*
		if len(poly.InteriorRings()) > 0 {
			return nil, errors.New("Polygon has interior rings")
		}
		*/
		
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

# go-whosonfirst-shapefile

Go package for working with Who's On First documents and ESRI shapefiles.

## Install

You will need to have both `Go` (specifically a version of Go more recent than 1.6 so let's just assume you need [Go 1.8](https://golang.org/dl/) or higher) and the `make` programs installed on your computer. Assuming you do just type:

```
make bin
```

All of this package's dependencies are bundled with the code in the `vendor` directory.

## Important

This works. Until it doesn't. Currently it only write points and polygons.

There is a lot still to do including moving code in to generic package (library) functions and sorting out common attributes.

## Tools

### wof-shapefile-index

```
$> ./bin/wof-shapefile-index -shapetype POLYGON -out test.shp -timings /usr/local/data/whosonfirst-data-constituency-us/
...time passes

$> du -h test.*
3.0M	test.dbf
401M	test.shp
 56K	test.shx
```

Which ends up looking something like this:

![](docs/images/20180815-constituencies.png)

## See also:

* https://github.com/jonas-p/go-shp
* https://www.esri.com/library/whitepapers/pdfs/shapefile.pdf
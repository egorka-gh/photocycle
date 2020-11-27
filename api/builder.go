package api

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/egorka-gh/photocycle"
	"github.com/spf13/cast"
)

//Builder builds photocycle models by json keys from database
type Builder struct {
	//json keys map by family
	jmap            map[int][]photocycle.JSONMap
	deliveryMapping map[int]map[int]photocycle.DeliveryTypeMapping
}

//BuildPackage builds photocycle.Package from raw
func (b *Builder) BuildPackage(source int, raw map[string]interface{}) (*photocycle.Package, error) {
	res := &photocycle.Package{}
	fields, ok := b.jmap[5]
	if !ok {
		return res, errors.New("Buider init error, has no fields for family 5")
	}
	//create & fill json vs object fields
	jm := make(map[string]interface{})
	for _, f := range fields {
		v, ok := deepSearch(raw, f.JSONKey)
		if !ok {
			continue
		}
		jm[f.Field] = v
	}
	j, _ := json.Marshal(jm)
	//fill target struct
	err := json.Unmarshal(j, res)
	if err != nil {
		return nil, err
	}
	res.Source = source
	//map site delivery id to photocycle id
	//translate delivery id
	m, ok := b.deliveryMapping[source]
	if ok {
		dm, ok := m[res.NativeDeliveryID]
		if ok {
			res.DeliveryID = dm.DeliveryType
		}
	}
	//TODO count orders num?
	//build prorerties
	fields, ok = b.jmap[6]
	if !ok {
		return res, errors.New("Buider init error, has no fields for family 6")
	}
	props := make([]photocycle.PackageProperty, 0, len(fields))
	for _, f := range fields {
		v, ok := deepSearch(raw, f.JSONKey)
		if !ok {
			if f.Field == "debt_sum" {
				v = "0"
			} else {
				continue
			}
		}
		if v == "" {
			continue
		}
		p := photocycle.PackageProperty{}
		p.Source = source
		p.PackageID = res.ID
		p.Property = f.Field
		p.Value = cast.ToString(v)
		props = append(props, p)
	}
	res.Properties = props

	//build barcodes
	bars := make([]photocycle.PackageBarcode, 0)
	v, ok := raw["boxes"]
	if ok {
		arr, ok := v.([]interface{})
		if ok {
			for _, im := range arr {
				m, ok := im.(map[string]interface{})
				if ok {
					barcode := cast.ToString(m["barcode"])
					if barcode != "" {
						b := photocycle.PackageBarcode{}
						b.Source = source
						b.PackageID = res.ID
						b.BarcodeType = 2
						b.Barcode = barcode
						b.BoxNumber = cast.ToInt(m["number"])
						bars = append(bars, b)
					}
				}
			}
		}
	}
	v, ok = raw["barcodes"]
	if ok {
		arr, ok := v.([]interface{})
		if ok {
			for _, im := range arr {
				m, ok := im.(map[string]interface{})
				if ok {
					barcode := cast.ToString(m["barcode"])
					if barcode != "" {
						b := photocycle.PackageBarcode{}
						b.Source = source
						b.PackageID = res.ID
						b.BarcodeType = 1
						b.Barcode = barcode
						b.BoxNumber = cast.ToInt(m["number"])
						bars = append(bars, b)
					}
				}
			}
		}
	}

	res.Barcodes = bars

	return res, nil
}

// deepSearch scans deep maps,
//following the key indexes point delemited
//
// In case intermediate keys do not existreturns nil, false
func deepSearch(m map[string]interface{}, key string) (interface{}, bool) {
	path := strings.Split(key, ".")
	lastKey := path[len(path)-1]
	path = path[0 : len(path)-1]
	for _, k := range path {
		m2, ok := m[k]
		if !ok {
			// intermediate key does not exist
			return nil, false
		}
		m3, ok := m2.(map[string]interface{})
		if !ok {
			// intermediate key is a value
			return nil, false
		}
		// continue search from here
		m = m3
	}
	v, ok := m[lastKey]
	return v, ok
}

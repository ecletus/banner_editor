// +build !bindatafs

package bindatafs

import "fmt"

var _bindata = map[string]interface{}{}

func Asset(name string) (*assetfs.Asset, error) {
	return nil, fmt.Errorf("Asset %s not found", name)
}

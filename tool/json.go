package tool

import "encoding/json"

func MustItfToJsonStr(v interface{}) string {
	if bts, err := json.Marshal(v); err != nil {
		panic(err)
	} else {
		return string(bts)
	}
}

func MustItfToJsonStrIndex(v interface{}) string {
	if bts, err := json.MarshalIndent(v, "", "  "); err != nil {
		panic(err)
	} else {
		return string(bts)
	}
}

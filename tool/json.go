package tool

import "encoding/json"

func MustItfToJSONStr(v any) string {
	bts, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(bts)
}

func MustItfToJSONStrIndex(v any) string {
	bts, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		panic(err)
	}
	return string(bts)
}

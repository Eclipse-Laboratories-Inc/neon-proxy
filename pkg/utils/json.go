package utils

import "encoding/json"

// JsonToStruct return populates the fields of the dst struct from the fields
// of the src struct using json tags
func JsonToStruct(src any, dst any) error {
	result, err := json.Marshal(src)
	if err != nil {
		return err
	}

	return json.Unmarshal(result, dst)
}

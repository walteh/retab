package buildrc

import "encoding/json"

func (me *BuildrcJSON) Files() (map[string]string, error) {
	ok, err := json.Marshal(me)
	if err != nil {
		return nil, err
	}

	var res map[string]any

	err = json.Unmarshal(ok, &res)
	if err != nil {
		return nil, err
	}

	ok2 := map[string]string{}
	for k, v := range res {
		if str, ok := v.(string); ok {
			ok2[k] = str
			continue
		}
		a, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}

		ok2[k] = string(a)
	}

	return ok2, nil
}

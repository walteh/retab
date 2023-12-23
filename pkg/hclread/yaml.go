package hclread

// type FlowMap struct {

// }

// type FlowValueInternal interface {
// 	// *FlowMapValueContent | *FlowMap
// }

// type FlowMapValueContent struct {
// 	Content any `yaml:",flow"`
// }

// type FlowMapMap struct {
// 	Content map[string]FlowValueInternal `yaml:",flow"`
// }

// func (me *FlowMap) MarshalYAML() (any, error) {
// 	wrt := bytes.NewBuffer(nil)
// 	enc := yaml.NewEncoder(wrt)
// 	enc.SetIndent(4)

// 	err := enc.Encode(me)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return wrt.String(), nil
// }

// func NewFlowMap(a any) FlowValueInternal {
// 	switch v := a.(type) {
// 	case map[string]any:
// 		val := make(map[string]FlowValueInternal)
// 		for k, vv := range v {
// 			val[k] = NewFlowMap(vv)
// 		}
// 		return &FlowMap{
// 			Content: val,
// 		}
// 	default:
// 		return &FlowMapValueContent{
// 			Content: a,
// 		}
// 	}
// }

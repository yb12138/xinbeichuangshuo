// gameflow: InterruptChoice 上下文 map 归一化为 map[string]any。

package engine

// choiceCtxAsAnyMap 将中断 Context 转为 map[string]any（map[string]interface{} 与 map[string]any 为同一类型）。
func choiceCtxAsAnyMap(raw interface{}) (map[string]any, bool) {
	m, ok := raw.(map[string]any)
	if !ok {
		return nil, false
	}
	return m, true
}

func choiceCtxAsInterfaceMap(m map[string]any) map[string]interface{} {
	if m == nil {
		return nil
	}
	out := make(map[string]interface{}, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

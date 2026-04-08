package engine

// skillChoiceInputHandler 表示一个 choice_type 对应的一步结算处理器。
// 返回 error 表示该步输入不合法或规则不成立；nil 表示该步处理成功。
type skillChoiceInputHandler func(selectionIndex int, ctxData map[string]interface{}) error

// dispatchChoiceInputByType 统一处理 skill choice 的路由分发。
// 命中 handler 时返回 (true, err)；未命中返回 (false, nil) 交由其它流程继续尝试。
func dispatchChoiceInputByType(choiceType string, selectionIndex int, ctxData map[string]interface{}, handlers map[string]skillChoiceInputHandler) (bool, error) {
	handler, ok := handlers[choiceType]
	if !ok {
		return false, nil
	}
	return true, handler(selectionIndex, ctxData)
}

// skillChoiceRouteHandler 用于已经实现了 (bool, error) 语义的旧式 choice handler。
type skillChoiceRouteHandler func(selectionIndex int, ctxData map[string]interface{}) (bool, error)

// dispatchChoiceRouteByType 保留旧式 handler 的返回语义，同时将入口统一为 map 分发。
func dispatchChoiceRouteByType(choiceType string, selectionIndex int, ctxData map[string]interface{}, handlers map[string]skillChoiceRouteHandler) (bool, error) {
	handler, ok := handlers[choiceType]
	if !ok {
		return false, nil
	}
	return handler(selectionIndex, ctxData)
}

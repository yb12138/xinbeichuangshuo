// gameflow: Choice 引擎宿主标记（具体 *GameEngine 由 engine.choiceHostBridge 实现）。

package choice

// Host 由 engine.choiceHostBridge 实现；Spec 回调仅用于承载桥接类型。
type Host interface {
	ChoiceEngineHost()
}

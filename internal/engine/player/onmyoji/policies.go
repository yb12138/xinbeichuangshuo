// gameflow: 鬼术师策略声明。

package onmyoji

// PolicySpecs 导出鬼术师策略类型声明（engine层根据此声明装配对应函数）。
var PolicySpecs = []string{
	"interaction_binding_interrupt",
	"interaction_binding_counter",
	"interaction_yinyang",
	"counter_element",
	"counter_resolve",
}

// gameflow: Timing 链式调用辅助。

package engine

// runTimingBoolChain 顺序执行同一 Timing 阶段的布尔钩子。
// 任何一个钩子返回 true 都会立即短路。

// runTimingErrorChain 顺序执行同一 Timing 阶段的校验策略。
// 任一策略返回错误即短路。
func runTimingErrorChain(steps ...func() error) error {
	for _, step := range steps {
		if step == nil {
			continue
		}
		if err := step(); err != nil {
			return err
		}
	}
	return nil
}

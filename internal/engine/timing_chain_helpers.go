package engine

// runTimingBoolChain 顺序执行同一 Timing 阶段的布尔钩子。
// 任何一个钩子返回 true 都会立即短路。
func runTimingBoolChain(steps ...func() bool) bool {
	for _, step := range steps {
		if step != nil && step() {
			return true
		}
	}
	return false
}

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

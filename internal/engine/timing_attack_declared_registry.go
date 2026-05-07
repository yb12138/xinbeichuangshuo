// gameflow: 动态钩子表重建。

package engine

// rebuildTimingOnAttackDeclaredRegistry 根据当前已上场角色，重建必要的执行表。
func (e *GameEngine) rebuildTimingOnAttackDeclaredRegistry() {
	if e == nil {
		return
	}

	// 通用流程钩子
	e.beforeActionFieldHooks = []turnTimingHook{
		beforeActionPoisonHook,
		beforeActionFiveElementsBindHook,
		beforeActionWeakHook,
	}

	// 游戏启动钩子
	e.gameStartAddPlayerHooks = []gameStartPlayerHook{bootstrapApplyRoleDefaults}
	e.gameStartInitialDealHooks = []gameStartPlayerHook{bootstrapEnsureStarterRoleCards}
}


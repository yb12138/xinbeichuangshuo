// gameflow: 技能 Handler 基础实现（从 engine/skill/registry.go 迁出）。

package player

import "starcup-engine/internal/model"

// BaseHandler 提供技能 Handler 的默认实现：CanUse 始终允许，Execute 为空操作。
type BaseHandler struct{}

func (h *BaseHandler) CanUse(ctx *model.Context) bool   { return true }
func (h *BaseHandler) Execute(ctx *model.Context) error { return nil }

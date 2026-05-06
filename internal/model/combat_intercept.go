package model

// CombatInterceptTag 表示当前战斗链上的即时劫持规则。
// 这些标记只描述“这次战斗怎么结算”，不承载跨窗口的角色资源状态。
type CombatInterceptTag string

const (
	CombatInterceptNone             CombatInterceptTag = ""
	CombatInterceptUnrespondable    CombatInterceptTag = "Unrespondable"
	CombatInterceptIgnoreHolyShield CombatInterceptTag = "IgnoreHolyShield"
	CombatInterceptForceHit         CombatInterceptTag = "ForceHit"
	CombatInterceptIgnoreTargetHoly CombatInterceptTag = "IgnoreTargetHoly"
)

func CloneCombatInterceptTags(src map[CombatInterceptTag]bool) map[CombatInterceptTag]bool {
	if len(src) == 0 {
		return nil
	}
	dst := make(map[CombatInterceptTag]bool, len(src))
	for tag, enabled := range src {
		if enabled {
			dst[tag] = true
		}
	}
	if len(dst) == 0 {
		return nil
	}
	return dst
}

func ensureCombatInterceptTags(tags *map[CombatInterceptTag]bool) map[CombatInterceptTag]bool {
	if *tags == nil {
		*tags = map[CombatInterceptTag]bool{}
	}
	return *tags
}

func (info *AttackEventInfo) SetInterceptTag(tag CombatInterceptTag) {
	if info == nil || tag == CombatInterceptNone {
		return
	}
	tags := ensureCombatInterceptTags(&info.InterceptTags)
	tags[tag] = true
	switch tag {
	case CombatInterceptForceHit:
		info.IsHitForced = true
		info.CanBeResponded = false
		// 机制语义：强制命中默认无视圣盾判定。
		info.IgnoreShield = true
		tags[CombatInterceptIgnoreHolyShield] = true
	case CombatInterceptUnrespondable:
		info.CanBeResponded = false
	case CombatInterceptIgnoreHolyShield:
		info.IgnoreShield = true
	}
}

func (info *AttackEventInfo) HasInterceptTag(tag CombatInterceptTag) bool {
	return info != nil && info.InterceptTags != nil && info.InterceptTags[tag]
}

func (req *CombatRequest) SetInterceptTag(tag CombatInterceptTag) {
	if req == nil || tag == CombatInterceptNone {
		return
	}
	tags := ensureCombatInterceptTags(&req.InterceptTags)
	tags[tag] = true
	switch tag {
	case CombatInterceptForceHit:
		req.IsForcedHit = true
		req.CanBeResponded = false
		// 机制语义：强制命中默认无视圣盾判定。
		req.IgnoreShield = true
		tags[CombatInterceptIgnoreHolyShield] = true
	case CombatInterceptUnrespondable:
		req.CanBeResponded = false
	case CombatInterceptIgnoreHolyShield:
		req.IgnoreShield = true
	}
}

func (req *CombatRequest) HasInterceptTag(tag CombatInterceptTag) bool {
	return req != nil && req.InterceptTags != nil && req.InterceptTags[tag]
}

func (pd *PendingDamage) SetInterceptTag(tag CombatInterceptTag) {
	if pd == nil || tag == CombatInterceptNone {
		return
	}
	tags := ensureCombatInterceptTags(&pd.InterceptTags)
	tags[tag] = true
	if tag == CombatInterceptForceHit {
		tags[CombatInterceptIgnoreHolyShield] = true
		pd.IgnoreShield = true
	}
	if tag == CombatInterceptIgnoreHolyShield {
		pd.IgnoreShield = true
	}
}

func (pd *PendingDamage) HasInterceptTag(tag CombatInterceptTag) bool {
	return pd != nil && pd.InterceptTags != nil && pd.InterceptTags[tag]
}

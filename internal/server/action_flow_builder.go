package server

import (
	"fmt"
	"sort"
	"strings"

	"starcup-engine/internal/model"
)

type actionFlowEdgeMeta struct {
	CombatID string
}

type actionFlowBuilder struct {
	flow                ActionFlowDTO
	displayName         func(string, string) string
	actorSeen           map[string]bool
	nameByPlayerID      map[string]string
	targetHintsByActor  map[string][]string
	edgeIndexByID       map[string]int
	edgeMetaByID        map[string]actionFlowEdgeMeta
	nodeIndexByID       map[string]int
	skillNodeIndexByKey map[string]int
	latestEdgeID        string
	latestPendingEdgeID string
	nodeSeq             int
	edgeSeq             int
}

func buildLiveActionFlows(history []TimelineEvent, event TimelineEvent, displayName func(string, string) string) []ActionFlowDTO {
	if event.ActionID == "" || event.NarrativeKind == "action_closed" {
		return nil
	}
	events := make([]TimelineEvent, 0, len(history)+1)
	for _, item := range history {
		if item.ActionID == event.ActionID {
			events = append(events, item)
		}
	}
	events = append(events, event)
	flow := buildActionFlow(events, displayName)
	if flow == nil {
		return nil
	}
	return []ActionFlowDTO{*flow}
}

func buildReplayActionFlows(events []TimelineEvent, displayName func(string, string) string) []ActionFlowDTO {
	type actionGroup struct {
		actionID string
		firstID  int64
		events   []TimelineEvent
	}

	groupByAction := map[string]*actionGroup{}
	order := make([]string, 0)
	for _, event := range events {
		if event.ActionID == "" || event.NarrativeKind == "action_closed" {
			continue
		}
		group := groupByAction[event.ActionID]
		if group == nil {
			group = &actionGroup{actionID: event.ActionID, firstID: event.EventID}
			groupByAction[event.ActionID] = group
			order = append(order, event.ActionID)
		}
		if group.firstID == 0 || event.EventID < group.firstID {
			group.firstID = event.EventID
		}
		group.events = append(group.events, event)
	}

	sort.SliceStable(order, func(i, j int) bool {
		return groupByAction[order[i]].firstID < groupByAction[order[j]].firstID
	})

	flows := make([]ActionFlowDTO, 0, len(order))
	for _, actionID := range order {
		flow := buildActionFlow(groupByAction[actionID].events, displayName)
		if flow != nil {
			flows = append(flows, *flow)
		}
	}
	return flows
}

func buildActionFlow(events []TimelineEvent, displayName func(string, string) string) *ActionFlowDTO {
	if len(events) == 0 {
		return nil
	}
	sort.SliceStable(events, func(i, j int) bool {
		return events[i].EventID < events[j].EventID
	})

	actionID := ""
	for _, event := range events {
		if event.ActionID != "" {
			actionID = event.ActionID
			break
		}
	}
	if actionID == "" {
		return nil
	}

	builder := &actionFlowBuilder{
		flow: ActionFlowDTO{
			FlowID:     actionID,
			ActionID:   actionID,
			ActionType: "unknown",
		},
		displayName:         displayName,
		actorSeen:           map[string]bool{},
		nameByPlayerID:      map[string]string{},
		targetHintsByActor:  map[string][]string{},
		edgeIndexByID:       map[string]int{},
		edgeMetaByID:        map[string]actionFlowEdgeMeta{},
		nodeIndexByID:       map[string]int{},
		skillNodeIndexByKey: map[string]int{},
	}

	for _, event := range events {
		builder.rememberEventMetadata(event)
	}
	for _, event := range events {
		builder.applyBackboneEvent(event)
	}
	for _, event := range events {
		builder.applySkillOrEffectEvent(event)
	}
	for _, event := range events {
		builder.applyExtraActionNodeEvent(event)
	}
	builder.ensureFallbackActionNode()
	builder.finalizeActors()
	builder.buildLogs()

	if len(builder.flow.Nodes) == 0 && len(builder.flow.Edges) == 0 {
		return nil
	}
	return &builder.flow
}

func (b *actionFlowBuilder) rememberEventMetadata(event TimelineEvent) {
	if event.NarrativeWindowID != "" && b.flow.NarrativeWindowID == "" {
		b.flow.NarrativeWindowID = event.NarrativeWindowID
	}
	if event.ActorUserID != "" && b.flow.ActorUserID == "" {
		b.flow.ActorUserID = event.ActorUserID
	}
	if nextActionType := normalizeFlowActionType(event); b.flow.ActionType == "unknown" && nextActionType != "unknown" {
		b.flow.ActionType = nextActionType
	}

	b.rememberName(event.ActorUserID, event.ActorName)
	if len(event.TargetUserIDs) == 1 {
		b.rememberName(event.TargetUserIDs[0], event.TargetName)
	}
	b.addActor(event.ActorUserID)
	for _, targetID := range event.TargetUserIDs {
		b.addActor(targetID)
	}
	if event.ActorUserID != "" && len(event.TargetUserIDs) > 0 && !isExtraActionFlowEvent(event) {
		b.targetHintsByActor[event.ActorUserID] = uniqueFlowStrings(event.TargetUserIDs)
	}
}

func (b *actionFlowBuilder) applyBackboneEvent(event TimelineEvent) {
	kind := strings.TrimSpace(event.NarrativeKind)
	visualKind := strings.TrimSpace(event.VisualKind)
	cardRole := normalizeFlowPhase(firstNonEmptyString(event.CardRole, event.ActionType))

	switch {
	case kind == "action_started":
		if event.ActorUserID != "" && b.flow.ActorUserID == "" {
			b.flow.ActorUserID = event.ActorUserID
		}
		return
	case kind == "combat_declared" || kind == "combat_response":
		phase := cardRole
		if phase == "" || phase == "unknown" {
			if kind == "combat_declared" {
				phase = "attack"
			} else {
				phase = normalizeFlowPhase(event.CuePhase)
			}
		}
		if phase == "" || phase == "unknown" {
			phase = "counter"
		}
		for _, targetID := range event.TargetUserIDs {
			b.ensureEdge(event.ActorUserID, targetID, phase, event.CombatID)
		}
		return
	case visualKind == "card" && len(event.Cards) > 0:
		b.applyCardEvent(event, cardRole)
		return
	case kind == "damage_dealt":
		b.applyDamageEvent(event)
		return
	case event.EffectType == "attack_miss":
		b.applyMissEvent(event)
		return
	}
}

func (b *actionFlowBuilder) applySkillOrEffectEvent(event TimelineEvent) {
	kind := strings.TrimSpace(event.NarrativeKind)
	visualKind := strings.TrimSpace(event.VisualKind)

	switch {
	case strings.HasPrefix(kind, "skill") || event.GameplayType == "skill_activated" || visualKind == "skill_token":
		b.applySkillEvent(event)
		return
	case kind == "field_effect_applied" || kind == "field_effect_removed" || visualKind == "effect_token":
		b.applyEffectEvent(event)
		return
	}
}

func (b *actionFlowBuilder) applyExtraActionNodeEvent(event TimelineEvent) {
	if isExtraActionFlowEvent(event) {
		b.applyExtraActionEvent(event)
	}
}

func (b *actionFlowBuilder) applyCardEvent(event TimelineEvent, phase string) {
	if isCombatFlowPhase(phase) || phase == "magic" {
		targets := b.targetsForCardEvent(event, phase)
		for _, targetID := range targets {
			edgeID := b.ensureEdge(event.ActorUserID, targetID, phase, event.CombatID)
			if edgeID == "" {
				continue
			}
			b.attachCardsToEdge(edgeID, event.EventID, event.Cards)
			nodeID := b.addNode(ActionFlowNodeDTO{
				Kind:          "card",
				ActorUserID:   event.ActorUserID,
				TargetUserIDs: []string{targetID},
				EventID:       event.EventID,
				AnchorEdgeID:  edgeID,
				Cards:         cloneFlowCards(event.Cards),
				Label:         firstCardName(event.Cards),
			})
			b.attachNodeToEdge(edgeID, nodeID)
		}
		if len(targets) == 0 {
			b.addNode(ActionFlowNodeDTO{
				Kind:        "card",
				ActorUserID: event.ActorUserID,
				EventID:     event.EventID,
				Cards:       cloneFlowCards(event.Cards),
				Label:       firstCardName(event.Cards),
			})
		}
		return
	}

	if phase == "field_effect" {
		targets := b.targetsForCardEvent(event, "effect")
		anchorEdgeID := ""
		if len(targets) > 0 {
			anchorEdgeID = b.ensureEdge(event.ActorUserID, targets[0], "effect", event.CombatID)
			b.attachCardsToEdge(anchorEdgeID, event.EventID, event.Cards)
		}
		nodeID := b.addNode(ActionFlowNodeDTO{
			Kind:          "effect",
			ActorUserID:   event.ActorUserID,
			TargetUserIDs: targets,
			EventID:       event.EventID,
			AnchorEdgeID:  anchorEdgeID,
			Cards:         cloneFlowCards(event.Cards),
			Label:         firstCardName(event.Cards),
			Outcome:       "resolved",
		})
		b.attachNodeToEdge(anchorEdgeID, nodeID)
	}
}

func (b *actionFlowBuilder) applySkillEvent(event TimelineEvent) {
	targets := uniqueFlowStrings(event.TargetUserIDs)
	anchorEdgeID := ""
	if b.flow.ActionType == "skill" && event.ActorUserID != "" && len(targets) > 0 {
		anchorEdgeID = b.ensureEdge(event.ActorUserID, targets[0], "skill", event.CombatID)
	} else {
		anchorEdgeID = b.findSkillTriggerEdge(event, targets)
	}
	if len(targets) == 0 && anchorEdgeID != "" {
		if edge := b.edgeByID(anchorEdgeID); edge != nil && edge.ToUserID != "" {
			targets = []string{edge.ToUserID}
		}
	}
	key := actionFlowSkillNodeKey(event, targets, anchorEdgeID)
	if index, ok := b.skillNodeIndexByKey[key]; ok && index >= 0 && index < len(b.flow.Nodes) {
		node := &b.flow.Nodes[index]
		if node.EventID == 0 {
			node.EventID = event.EventID
		}
		if node.AnchorEdgeID == "" {
			node.AnchorEdgeID = anchorEdgeID
		}
		if node.SkillID == "" {
			node.SkillID = event.SkillID
		}
		if node.SkillName == "" {
			node.SkillName = firstNonEmptyString(event.SkillName, event.Summary)
		}
		if node.Label == "" {
			node.Label = firstNonEmptyString(event.SkillName, event.Summary, "技能")
		}
		if len(node.TargetUserIDs) == 0 {
			node.TargetUserIDs = targets
		}
		b.attachNodeToEdge(anchorEdgeID, node.ID)
		return
	}
	nodeID := b.addNode(ActionFlowNodeDTO{
		Kind:          "skill",
		ActorUserID:   event.ActorUserID,
		TargetUserIDs: targets,
		EventID:       event.EventID,
		AnchorEdgeID:  anchorEdgeID,
		SkillID:       event.SkillID,
		SkillName:     firstNonEmptyString(event.SkillName, event.Summary),
		EffectText:    event.EffectText,
		Outcome:       "resolved",
		Label:         firstNonEmptyString(event.SkillName, event.Summary, "技能"),
	})
	if index, ok := b.nodeIndexByID[nodeID]; ok {
		b.skillNodeIndexByKey[key] = index
	}
	b.attachNodeToEdge(anchorEdgeID, nodeID)
}

func actionFlowSkillNodeKey(event TimelineEvent, targets []string, anchorEdgeID string) string {
	sortedTargets := append([]string{}, uniqueFlowStrings(targets)...)
	sort.Strings(sortedTargets)
	skillKey := firstNonEmptyString(event.SkillID, event.SkillName, event.Summary, "技能")
	return strings.Join([]string{
		strings.TrimSpace(event.ActorUserID),
		strings.TrimSpace(skillKey),
		strings.TrimSpace(anchorEdgeID),
		strings.Join(sortedTargets, ","),
	}, "|")
}

func (b *actionFlowBuilder) findSkillTriggerEdge(event TimelineEvent, targets []string) string {
	targets = uniqueFlowStrings(targets)
	if len(targets) == 0 {
		targets = uniqueFlowStrings(b.targetHintsByActor[event.ActorUserID])
	}

	if event.ActorUserID != "" && len(targets) > 0 {
		for _, phase := range preferredSkillAnchorPhases(b.flow.ActionType) {
			if edgeID := b.findEdgeByActorTargetsPhase(event.ActorUserID, targets, phase, event.CombatID); edgeID != "" {
				return edgeID
			}
		}
		for _, phase := range []string{"attack", "counter", "magic", "effect"} {
			if edgeID := b.findEdgeByActorTargetsPhase(event.ActorUserID, targets, phase, event.CombatID); edgeID != "" {
				return edgeID
			}
		}
	}

	if event.CombatID != "" && event.ActorUserID != "" {
		for _, phase := range []string{"attack", "counter", "magic", "effect"} {
			if edgeID := b.findEdgeByActorPhase(event.ActorUserID, phase, event.CombatID); edgeID != "" {
				return edgeID
			}
		}
	}

	return b.findAnchorEdge(event.ActorUserID, targets, event.CombatID)
}

func preferredSkillAnchorPhases(actionType string) []string {
	switch normalizeFlowPhase(actionType) {
	case "attack":
		return []string{"attack", "counter"}
	case "magic":
		return []string{"magic", "effect"}
	case "special":
		return []string{"effect", "skill"}
	default:
		return []string{"attack", "counter", "magic", "effect"}
	}
}

func (b *actionFlowBuilder) applyEffectEvent(event TimelineEvent) {
	targets := uniqueFlowStrings(event.TargetUserIDs)
	anchorEdgeID := b.findAnchorEdge(event.ActorUserID, targets, event.CombatID)
	if anchorEdgeID == "" && event.ActorUserID != "" && len(targets) > 0 && (b.flow.ActionType == "skill" || b.flow.ActionType == "magic") {
		anchorEdgeID = b.ensureEdge(event.ActorUserID, targets[0], "effect", event.CombatID)
	}
	if anchorEdgeID != "" {
		b.updateEdge(anchorEdgeID, func(edge *ActionFlowEdgeDTO) {
			if edge.Outcome == "" || edge.Outcome == "pending" {
				edge.Outcome = "resolved"
			}
			if edge.Label == "" {
				edge.Label = firstNonEmptyString(timelineEventFieldCardName(event), event.Summary, event.EffectType)
			}
		})
		return
	}
	label := firstNonEmptyString(timelineEventFieldCardName(event), event.Summary, event.EffectType, "效果")
	nodeID := b.addNode(ActionFlowNodeDTO{
		Kind:          "effect",
		ActorUserID:   event.ActorUserID,
		TargetUserIDs: targets,
		EventID:       event.EventID,
		AnchorEdgeID:  anchorEdgeID,
		Outcome:       "resolved",
		Label:         label,
	})
	b.attachNodeToEdge(anchorEdgeID, nodeID)
}

func (b *actionFlowBuilder) applyDamageEvent(event TimelineEvent) {
	targets := uniqueFlowStrings(event.TargetUserIDs)
	if len(targets) == 0 {
		return
	}
	for _, targetID := range targets {
		edgeID := b.findEdgeForResolution(event.ActorUserID, targetID, event.CombatID)
		if edgeID == "" {
			edgeID = b.ensureEdge(event.ActorUserID, targetID, "take", event.CombatID)
		}
		b.updateEdge(edgeID, func(edge *ActionFlowEdgeDTO) {
			edge.Damage = event.Damage
			edge.DamageType = event.DamageType
			edge.Outcome = "hit"
		})
	}
}

func (b *actionFlowBuilder) applyMissEvent(event TimelineEvent) {
	targets := uniqueFlowStrings(event.TargetUserIDs)
	if len(targets) == 0 {
		return
	}
	for _, targetID := range targets {
		edgeID := b.findEdgeForResolution(event.ActorUserID, targetID, event.CombatID)
		if edgeID == "" {
			edgeID = b.ensureEdge(event.ActorUserID, targetID, "attack", event.CombatID)
		}
		b.updateEdge(edgeID, func(edge *ActionFlowEdgeDTO) {
			edge.Outcome = "miss"
			edge.Label = "未命中"
		})
	}
}

func (b *actionFlowBuilder) applyExtraActionEvent(event TimelineEvent) {
	label := firstNonEmptyString(event.Summary, event.ExtraActionType, "额外行动")
	if event.ExtraActionElement != "" {
		label = fmt.Sprintf("%s %s", label, event.ExtraActionElement)
	}
	detail := extraActionDetailText(event)
	if b.mergeExtraActionIntoSkillNode(event, detail) {
		return
	}
	b.addNode(ActionFlowNodeDTO{
		Kind:        "extra_action",
		ActorUserID: event.ActorUserID,
		EventID:     event.EventID,
		EffectText:  detail,
		Label:       label,
		Outcome:     "resolved",
	})
}

func (b *actionFlowBuilder) mergeExtraActionIntoSkillNode(event TimelineEvent, detail string) bool {
	sourceName := strings.TrimSpace(event.Summary)
	if sourceName == "" || event.ActorUserID == "" {
		return false
	}
	for index := len(b.flow.Nodes) - 1; index >= 0; index-- {
		node := &b.flow.Nodes[index]
		if node.Kind != "skill" || node.ActorUserID != event.ActorUserID {
			continue
		}
		if !skillNodeMatchesExtraActionSource(*node, sourceName) {
			continue
		}
		node.EffectText = appendFlowDetail(node.EffectText, detail)
		if node.Label == "" {
			node.Label = firstNonEmptyString(node.SkillName, sourceName)
		}
		return true
	}
	return false
}

func (b *actionFlowBuilder) targetsForCardEvent(event TimelineEvent, phase string) []string {
	targets := uniqueFlowStrings(event.TargetUserIDs)
	if len(targets) > 0 {
		return targets
	}
	if edgeID := b.findEdgeByActorPhase(event.ActorUserID, phase, event.CombatID); edgeID != "" {
		edge := b.edgeByID(edgeID)
		if edge != nil && edge.ToUserID != "" {
			return []string{edge.ToUserID}
		}
	}
	if hints := b.targetHintsByActor[event.ActorUserID]; len(hints) > 0 {
		return append([]string{}, hints...)
	}
	return nil
}

func (b *actionFlowBuilder) ensureEdge(fromUserID, toUserID, phase, combatID string) string {
	fromUserID = strings.TrimSpace(fromUserID)
	toUserID = strings.TrimSpace(toUserID)
	phase = normalizeFlowPhase(phase)
	if fromUserID == "" || toUserID == "" || phase == "" || phase == "unknown" {
		return ""
	}
	if edgeID := b.findEdge(fromUserID, toUserID, phase, combatID); edgeID != "" {
		b.latestEdgeID = edgeID
		if edge := b.edgeByID(edgeID); edge != nil && edge.Outcome == "pending" {
			b.latestPendingEdgeID = edgeID
		}
		return edgeID
	}

	b.edgeSeq++
	edgeID := fmt.Sprintf("%s-edge-%d", b.flow.ActionID, b.edgeSeq)
	outcome := "pending"
	if phase == "defend" || phase == "shield" {
		outcome = "resolved"
	}
	edge := ActionFlowEdgeDTO{
		ID:         edgeID,
		Order:      b.edgeSeq,
		FromUserID: fromUserID,
		ToUserID:   toUserID,
		Phase:      phase,
		Outcome:    outcome,
	}
	b.flow.Edges = append(b.flow.Edges, edge)
	b.edgeIndexByID[edgeID] = len(b.flow.Edges) - 1
	b.edgeMetaByID[edgeID] = actionFlowEdgeMeta{CombatID: combatID}
	b.latestEdgeID = edgeID
	if outcome == "pending" {
		b.latestPendingEdgeID = edgeID
	}
	b.addActor(fromUserID)
	b.addActor(toUserID)
	return edgeID
}

func (b *actionFlowBuilder) findEdge(fromUserID, toUserID, phase, combatID string) string {
	for i := len(b.flow.Edges) - 1; i >= 0; i-- {
		edge := b.flow.Edges[i]
		if edge.FromUserID != fromUserID || edge.ToUserID != toUserID || edge.Phase != phase {
			continue
		}
		meta := b.edgeMetaByID[edge.ID]
		if combatID == "" || meta.CombatID == "" || meta.CombatID == combatID {
			return edge.ID
		}
	}
	return ""
}

func (b *actionFlowBuilder) findEdgeByActorPhase(actorID, phase, combatID string) string {
	for i := len(b.flow.Edges) - 1; i >= 0; i-- {
		edge := b.flow.Edges[i]
		if edge.FromUserID != actorID || edge.Phase != phase {
			continue
		}
		meta := b.edgeMetaByID[edge.ID]
		if combatID == "" || meta.CombatID == "" || meta.CombatID == combatID {
			return edge.ID
		}
	}
	return ""
}

func (b *actionFlowBuilder) findEdgeByActorTargetsPhase(actorID string, targetIDs []string, phase, combatID string) string {
	targetSet := map[string]bool{}
	for _, targetID := range uniqueFlowStrings(targetIDs) {
		targetSet[targetID] = true
	}
	if actorID == "" || len(targetSet) == 0 {
		return ""
	}
	for i := len(b.flow.Edges) - 1; i >= 0; i-- {
		edge := b.flow.Edges[i]
		if edge.FromUserID != actorID || !targetSet[edge.ToUserID] || edge.Phase != phase {
			continue
		}
		meta := b.edgeMetaByID[edge.ID]
		if combatID == "" || meta.CombatID == "" || meta.CombatID == combatID {
			return edge.ID
		}
	}
	return ""
}

func (b *actionFlowBuilder) findEdgeForResolution(actorID, targetID, combatID string) string {
	if combatID != "" {
		for i := len(b.flow.Edges) - 1; i >= 0; i-- {
			edge := b.flow.Edges[i]
			if edge.FromUserID == actorID && edge.ToUserID == targetID && b.edgeMetaByID[edge.ID].CombatID == combatID {
				return edge.ID
			}
		}
	}
	for i := len(b.flow.Edges) - 1; i >= 0; i-- {
		edge := b.flow.Edges[i]
		if edge.FromUserID == actorID && edge.ToUserID == targetID {
			return edge.ID
		}
	}
	if b.latestPendingEdgeID != "" {
		return b.latestPendingEdgeID
	}
	return b.latestEdgeID
}

func (b *actionFlowBuilder) findAnchorEdge(actorID string, targetIDs []string, combatID string) string {
	if actorID != "" {
		for _, targetID := range targetIDs {
			if edgeID := b.findEdgeForResolution(actorID, targetID, combatID); edgeID != "" {
				return edgeID
			}
		}
	}
	if b.latestPendingEdgeID != "" {
		return b.latestPendingEdgeID
	}
	return b.latestEdgeID
}

func (b *actionFlowBuilder) attachCardsToEdge(edgeID string, eventID int64, cards []model.Card) {
	if edgeID == "" || len(cards) == 0 {
		return
	}
	b.updateEdge(edgeID, func(edge *ActionFlowEdgeDTO) {
		if edge.CardEventID == 0 {
			edge.CardEventID = eventID
		}
		edge.Cards = appendUniqueFlowCards(edge.Cards, cards)
	})
}

func (b *actionFlowBuilder) addNode(node ActionFlowNodeDTO) string {
	b.nodeSeq++
	node.ID = fmt.Sprintf("%s-node-%d", b.flow.ActionID, b.nodeSeq)
	node.Order = b.nodeSeq
	node.TargetUserIDs = uniqueFlowStrings(node.TargetUserIDs)
	node.Cards = cloneFlowCards(node.Cards)
	b.flow.Nodes = append(b.flow.Nodes, node)
	b.nodeIndexByID[node.ID] = len(b.flow.Nodes) - 1
	b.addActor(node.ActorUserID)
	for _, targetID := range node.TargetUserIDs {
		b.addActor(targetID)
	}
	return node.ID
}

func (b *actionFlowBuilder) attachNodeToEdge(edgeID, nodeID string) {
	if edgeID == "" || nodeID == "" {
		return
	}
	b.updateEdge(edgeID, func(edge *ActionFlowEdgeDTO) {
		if !containsFlowString(edge.NodeIDs, nodeID) {
			edge.NodeIDs = append(edge.NodeIDs, nodeID)
		}
	})
	if nodeIndex, ok := b.nodeIndexByID[nodeID]; ok && b.flow.Nodes[nodeIndex].AnchorEdgeID == "" {
		b.flow.Nodes[nodeIndex].AnchorEdgeID = edgeID
	}
}

func (b *actionFlowBuilder) updateEdge(edgeID string, update func(*ActionFlowEdgeDTO)) {
	index, ok := b.edgeIndexByID[edgeID]
	if !ok || index < 0 || index >= len(b.flow.Edges) || update == nil {
		return
	}
	update(&b.flow.Edges[index])
}

func (b *actionFlowBuilder) edgeByID(edgeID string) *ActionFlowEdgeDTO {
	index, ok := b.edgeIndexByID[edgeID]
	if !ok || index < 0 || index >= len(b.flow.Edges) {
		return nil
	}
	return &b.flow.Edges[index]
}

func (b *actionFlowBuilder) ensureFallbackActionNode() {
	if len(b.flow.Nodes) > 0 || len(b.flow.Edges) > 0 || b.flow.ActorUserID == "" {
		return
	}
	b.addNode(ActionFlowNodeDTO{
		Kind:        "action",
		ActorUserID: b.flow.ActorUserID,
		Label:       actionTypeVerb(b.flow.ActionType),
		Outcome:     "pending",
	})
}

func (b *actionFlowBuilder) addActor(playerID string) {
	playerID = strings.TrimSpace(playerID)
	if playerID == "" || b.actorSeen[playerID] {
		return
	}
	b.actorSeen[playerID] = true
	b.flow.Actors = append(b.flow.Actors, ActionFlowActorDTO{
		PlayerID: playerID,
		Order:    len(b.flow.Actors) + 1,
	})
}

func (b *actionFlowBuilder) rememberName(playerID, name string) {
	playerID = strings.TrimSpace(playerID)
	name = strings.TrimSpace(name)
	if playerID == "" || name == "" || b.nameByPlayerID[playerID] != "" {
		return
	}
	b.nameByPlayerID[playerID] = name
}

func (b *actionFlowBuilder) nameFor(playerID string) string {
	fallback := firstNonEmptyString(b.nameByPlayerID[playerID], playerID)
	if b.displayName != nil {
		return b.displayName(playerID, fallback)
	}
	return fallback
}

func (b *actionFlowBuilder) finalizeActors() {
	if b.flow.ActorUserID != "" {
		b.addActor(b.flow.ActorUserID)
	}
	for index := range b.flow.Actors {
		b.flow.Actors[index].Order = index + 1
	}
}

func (b *actionFlowBuilder) buildLogs() {
	type logCandidate struct {
		order int
		text  string
	}
	candidates := make([]logCandidate, 0, len(b.flow.Edges)+len(b.flow.Nodes))
	for _, edge := range b.flow.Edges {
		parts := []string{
			actionTypeVerb(edge.Phase),
			fmt.Sprintf("%s -> %s", b.nameFor(edge.FromUserID), b.nameFor(edge.ToUserID)),
		}
		if cardName := firstCardName(edge.Cards); cardName != "" {
			parts = append(parts, cardName)
		}
		if edge.Outcome == "miss" || edge.Label == "未命中" {
			parts = append(parts, "未命中")
		} else if edge.Damage > 0 {
			parts = append(parts, fmt.Sprintf("伤害: %d", edge.Damage))
		} else if edge.Outcome == "resolved" {
			parts = append(parts, "已结算")
		}
		candidates = append(candidates, logCandidate{
			order: edge.Order * 10,
			text:  strings.Join(parts, " | "),
		})
	}
	for _, node := range b.flow.Nodes {
		if node.Kind != "skill" && node.Kind != "extra_action" && node.Kind != "effect" {
			continue
		}
		if node.AnchorEdgeID != "" && node.Kind == "effect" {
			continue
		}
		verb := actionTypeVerb(node.Kind)
		label := firstNonEmptyString(node.SkillName, node.Label, node.EffectText)
		if label == "" {
			continue
		}
		actor := b.nameFor(node.ActorUserID)
		text := fmt.Sprintf("%s %s", verb, actor)
		if actor == "" {
			text = verb
		}
		text = strings.TrimSpace(text + "：" + label)
		candidates = append(candidates, logCandidate{
			order: 10000 + node.Order,
			text:  text,
		})
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		return candidates[i].order < candidates[j].order
	})
	b.flow.Logs = make([]ActionFlowLogDTO, 0, len(candidates))
	for index, candidate := range candidates {
		b.flow.Logs = append(b.flow.Logs, ActionFlowLogDTO{
			Order: index + 1,
			Text:  candidate.text,
		})
	}
}

func isExtraActionFlowEvent(event TimelineEvent) bool {
	kind := strings.TrimSpace(event.NarrativeKind)
	return kind == "extra_action_granted" || strings.TrimSpace(event.ExtraActionType) != ""
}

func skillNodeMatchesExtraActionSource(node ActionFlowNodeDTO, sourceName string) bool {
	sourceName = strings.TrimSpace(sourceName)
	if sourceName == "" {
		return false
	}
	for _, candidate := range []string{node.SkillName, node.Label} {
		if strings.TrimSpace(candidate) == sourceName {
			return true
		}
	}
	return false
}

func appendFlowDetail(existing, detail string) string {
	existing = strings.TrimSpace(existing)
	detail = strings.TrimSpace(detail)
	if detail == "" || strings.Contains(existing, detail) {
		return existing
	}
	if existing == "" {
		return detail
	}
	return existing + "；" + detail
}

func extraActionDetailText(event TimelineEvent) string {
	actionText := extraActionTypeText(event.ExtraActionType)
	elementText := extraActionElementText(event.ExtraActionElement)
	return "额外+1" + elementText + actionText
}

func extraActionTypeText(value string) string {
	switch normalizeFlowPhase(value) {
	case "attack":
		return "攻击行动"
	case "magic":
		return "法术行动"
	case "special":
		return "特殊行动"
	case "", strings.ToLower(model.ExtraActionAny):
		return "行动"
	default:
		return "行动"
	}
}

func extraActionElementText(value string) string {
	switch model.Element(strings.TrimSpace(value)) {
	case model.ElementWater:
		return "水系"
	case model.ElementFire:
		return "火系"
	case model.ElementEarth:
		return "地系"
	case model.ElementWind:
		return "风系"
	case model.ElementThunder:
		return "雷系"
	case model.ElementLight:
		return "光系"
	case model.ElementDark:
		return "暗系"
	default:
		return ""
	}
}

func normalizeFlowActionType(event TimelineEvent) string {
	for _, raw := range []string{event.ActionType, event.CardRole} {
		switch normalizeFlowPhase(raw) {
		case "attack":
			return "attack"
		case "magic":
			return "magic"
		case "skill":
			return "skill"
		}
	}
	kind := strings.TrimSpace(event.NarrativeKind)
	if strings.HasPrefix(kind, "skill") {
		if event.ActionID != "" && strings.HasSuffix(strings.ToLower(event.ActionID), "-skill") {
			return "skill"
		}
	}
	actionID := strings.ToLower(event.ActionID)
	for _, suffix := range []string{"attack", "magic", "skill", "special"} {
		if strings.HasSuffix(actionID, "-"+suffix) {
			return suffix
		}
	}
	return "unknown"
}

func normalizeFlowPhase(value string) string {
	normalized := strings.ToLower(strings.TrimSpace(value))
	switch normalized {
	case "attack":
		return "attack"
	case "counter", "respond":
		return "counter"
	case "defend", "defense":
		return "defend"
	case "shield":
		return "shield"
	case "magic":
		return "magic"
	case "skill":
		return "skill"
	case "field_effect", "effect":
		return "effect"
	case "take", "damage":
		return "take"
	case "buy", "synthesize", "extract", "special":
		return "special"
	case "extra_action":
		return "extra_action"
	case "":
		return ""
	default:
		return normalized
	}
}

func isCombatFlowPhase(phase string) bool {
	switch phase {
	case "attack", "counter", "defend", "shield":
		return true
	default:
		return false
	}
}

func actionTypeVerb(value string) string {
	switch normalizeFlowPhase(value) {
	case "attack":
		return "攻击"
	case "counter":
		return "应战"
	case "defend":
		return "防御"
	case "shield":
		return "圣盾"
	case "magic":
		return "法术"
	case "skill":
		return "技能"
	case "effect":
		return "效果"
	case "take":
		return "承受"
	case "extra_action":
		return "额外行动"
	case "special":
		return "特殊"
	default:
		return "行动"
	}
}

func cloneFlowCards(cards []model.Card) []model.Card {
	if len(cards) == 0 {
		return nil
	}
	out := make([]model.Card, len(cards))
	copy(out, cards)
	return out
}

func appendUniqueFlowCards(existing []model.Card, cards []model.Card) []model.Card {
	out := cloneFlowCards(existing)
	for _, card := range cards {
		key := flowCardKey(card)
		if key == "" {
			continue
		}
		exists := false
		for _, item := range out {
			if flowCardKey(item) == key {
				exists = true
				break
			}
		}
		if !exists {
			out = append(out, card)
		}
	}
	return out
}

func flowCardKey(card model.Card) string {
	if card.ID != "" {
		return "id:" + card.ID
	}
	return fmt.Sprintf("card:%s:%s:%s", card.Name, card.Type, card.Element)
}

func firstCardName(cards []model.Card) string {
	for _, card := range cards {
		if strings.TrimSpace(card.Name) != "" {
			return card.Name
		}
	}
	return ""
}

func uniqueFlowStrings(values []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	return out
}

func containsFlowString(values []string, needle string) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func timelineEventFieldCardName(event TimelineEvent) string {
	if event.FieldCard == nil {
		return ""
	}
	return event.FieldCard.Card.Name
}

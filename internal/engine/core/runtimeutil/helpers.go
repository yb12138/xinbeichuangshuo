package runtimeutil

import (
	"starcup-engine/internal/model"
	"strings"
)

func ToIntContextValue(v interface{}) int {
	if i, ok := v.(int); ok {
		return i
	}
	if f, ok := v.(float64); ok {
		return int(f)
	}
	return 0
}

func ToBoolContextValue(v interface{}) bool {
	if b, ok := v.(bool); ok {
		return b
	}
	return false
}

func IsMagicDamageType(damageType model.DamageType) bool {
	return !strings.EqualFold(string(damageType), "Attack")
}

func ParseStringSliceContextValue(v interface{}) []string {
	var out []string
	switch arr := v.(type) {
	case []string:
		out = append(out, arr...)
	case []interface{}:
		for _, item := range arr {
			if s, ok := item.(string); ok {
				out = append(out, s)
			}
		}
	}
	return out
}

func DedupeIDs(ids []string) []string {
	if len(ids) == 0 {
		return nil
	}
	out := make([]string, 0, len(ids))
	seen := map[string]bool{}
	for _, id := range ids {
		if id == "" || seen[id] {
			continue
		}
		seen[id] = true
		out = append(out, id)
	}
	return out
}

func IDsToSet(ids []string) map[string]bool {
	set := make(map[string]bool, len(ids))
	for _, id := range ids {
		if id == "" {
			continue
		}
		set[id] = true
	}
	return set
}

func ParseChoiceIntSlice(raw interface{}) []int {
	var out []int
	if arr, ok := raw.([]int); ok {
		out = append(out, arr...)
		return out
	}
	if arr, ok := raw.([]interface{}); ok {
		for _, v := range arr {
			if f, ok := v.(float64); ok {
				out = append(out, int(f))
			}
		}
	}
	return out
}

func ResolveSelectionToAllowedIndex(selection int, candidates []int, allowed map[int]struct{}) (int, bool) {
	if _, ok := allowed[selection]; ok {
		return selection, true
	}
	if selection >= 0 && selection < len(candidates) {
		mapped := candidates[selection]
		if _, ok := allowed[mapped]; ok {
			return mapped, true
		}
	}
	return 0, false
}

func ResolveSelectionToCandidate(selection int, candidates []int) (int, bool) {
	for _, candidate := range candidates {
		if selection == candidate {
			return candidate, true
		}
	}
	if selection >= 0 && selection < len(candidates) {
		return candidates[selection], true
	}
	return 0, false
}

func PickKIndices(src []int, k int) [][]int {
	var out [][]int
	var dfs func(start int, cur []int)
	dfs = func(start int, cur []int) {
		if len(cur) == k {
			cp := append([]int{}, cur...)
			out = append(out, cp)
			return
		}
		for i := start; i < len(src); i++ {
			cur = append(cur, src[i])
			dfs(i+1, cur)
			cur = cur[:len(cur)-1]
		}
	}
	dfs(0, nil)
	return out
}

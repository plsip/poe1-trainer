package rule

import (
	_ "embed"
	"encoding/json"
	"sync"
)

//go:embed seeds/stormburst_campaign.json
var storeburst1JSON []byte

var (
	once        sync.Once
	rulesBySlug map[string][]Rule
)

func initRules() {
	rulesBySlug = make(map[string][]Rule)
	for _, raw := range [][]byte{storeburst1JSON} {
		var f RulesFile
		if err := json.Unmarshal(raw, &f); err != nil {
			continue
		}
		rulesBySlug[f.GuideSlug] = append(rulesBySlug[f.GuideSlug], f.Rules...)
	}
}

// Engine evaluates embedded rule definitions against run state.
// It is stateless and safe for concurrent use after NewEngine returns.
type Engine struct{}

// NewEngine initialises the Engine and ensures all seed files are parsed.
func NewEngine() *Engine {
	once.Do(initRules)
	return &Engine{}
}

// Evaluate returns all active alerts for the given guide slug and context.
// Rules are returned in the order they appear in the seed file.
// Returns nil when no rules are registered for the guide slug.
func (e *Engine) Evaluate(guideSlug string, ctx EvalContext) []Alert {
	once.Do(initRules)
	rules := rulesBySlug[guideSlug]
	if len(rules) == 0 {
		return nil
	}
	var alerts []Alert
	for _, r := range rules {
		if conditionMet(r.Condition, ctx) {
			alerts = append(alerts, r.Alert)
		}
	}
	return alerts
}

// conditionMet returns true when every constrained field in c is satisfied by ctx.
// Fields with a zero value impose no restriction.
func conditionMet(c Condition, ctx EvalContext) bool {
	if c.MinAct > 0 && ctx.Act < c.MinAct {
		return false
	}
	if c.MaxAct > 0 && ctx.Act > c.MaxAct {
		return false
	}
	if c.MinLevel > 0 && ctx.Level > 0 && ctx.Level < c.MinLevel {
		return false
	}
	if c.MaxLevel > 0 && ctx.Level > 0 && ctx.Level > c.MaxLevel {
		return false
	}
	if len(c.StepTypes) > 0 {
		matched := false
		for _, t := range c.StepTypes {
			if t == ctx.StepType {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}
	return true
}

package model

import (
	"testing"
)

func TestListToDTO(t *testing.T) {
	// Case 1: Rules populated, RuleCount 0 (legacy/manual)
	l1 := List{
		Rules: []Rule{{ID: 1}, {ID: 2}},
	}
	dto1 := l1.ToDTO()
	if dto1.Rules != 2 {
		t.Errorf("Expected 2 rules, got %d", dto1.Rules)
	}

	// Case 2: Rules nil, RuleCount populated (optimized)
	l2 := List{
		Rules:     nil,
		RuleCount: 5,
	}
	dto2 := l2.ToDTO()
	if dto2.Rules != 5 {
		t.Errorf("Expected 5 rules, got %d", dto2.Rules)
	}

	// Case 3: Rules empty slice, RuleCount populated (optimized but empty slice initialized)
	l3 := List{
		Rules:     []Rule{},
		RuleCount: 3,
	}
	dto3 := l3.ToDTO()
	if dto3.Rules != 3 {
		t.Errorf("Expected 3 rules, got %d", dto3.Rules)
	}

	// Case 4: Rules populated, RuleCount populated (should prefer len(Rules) if consistent, or just match)
	l4 := List{
		Rules:     []Rule{{ID: 1}},
		RuleCount: 1,
	}
	dto4 := l4.ToDTO()
	if dto4.Rules != 1 {
		t.Errorf("Expected 1 rule, got %d", dto4.Rules)
	}

    // Case 5: Rules populated with DIFFERENT count than RuleCount (inconsistent state)
    // The code `if len(l.Rules) > 0 { rules = len(l.Rules) }` prefers actual rules.
    l5 := List{
        Rules: []Rule{{ID: 1}, {ID: 2}},
        RuleCount: 10,
    }
    dto5 := l5.ToDTO()
    if dto5.Rules != 2 {
        t.Errorf("Expected 2 rules (from slice), got %d", dto5.Rules)
    }
}

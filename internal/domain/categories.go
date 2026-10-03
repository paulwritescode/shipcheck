package domain

// ComputeCategoryStates returns the per-category completion summary for all
// seven categories, in the fixed Categories order (Req 7.6).
//
// For each category it reports the count of complete items and the total count
// of items, and classifies the category as:
//   - empty    when it contains zero items,
//   - complete when every item is complete for scoring,
//   - partial  when at least one but not all items are complete.
func ComputeCategoryStates(l LaunchData) []CategoryCompletion {
	// Tally complete/total per category in a single pass.
	type tally struct{ complete, total int }
	counts := make(map[Category]tally, len(Categories))
	for _, item := range l.ChecklistItems {
		t := counts[item.Category]
		t.total++
		if itemComplete(item) {
			t.complete++
		}
		counts[item.Category] = t
	}

	out := make([]CategoryCompletion, 0, len(Categories))
	for _, cat := range Categories {
		t := counts[cat]
		state := CategoryEmpty
		switch {
		case t.total == 0:
			state = CategoryEmpty
		case t.complete == t.total:
			state = CategoryComplete
		default:
			state = CategoryPartial
		}
		out = append(out, CategoryCompletion{
			Category:      cat,
			CompleteCount: t.complete,
			TotalCount:    t.total,
			State:         state,
		})
	}
	return out
}

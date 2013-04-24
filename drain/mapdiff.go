package drain

import (
	"sort"
)

type MapDiffChange struct {
	Deleted  bool
	Key      string
	NewValue string
	OldValue string
}

// MapDiff returns the difference between two maps (of
// string->string).
func MapDiff(oldMap, newMap map[string]string) []MapDiffChange {
	changes := []MapDiffChange{}
	oldKeys := mapSortedKeys(oldMap)
	newKeys := mapSortedKeys(newMap)

	i, j := 0, 0

	for i < len(oldKeys) && j < len(newKeys) {
		key1, key2 := oldKeys[i], newKeys[j]
		if key1 == key2 {
			val1, val2 := oldMap[key1], newMap[key2]
			if val1 != val2 {
				// key1 was changed.
				changes = append(changes, MapDiffChange{
					Key:      key1,
					OldValue: val1,
					NewValue: val2})
			}
			i++
			j++
		} else if key1 < key2 {
			// key1 was deleted.
			changes = append(changes, MapDiffChange{
				Deleted: true,
				Key:     key1})
			i++
		} else {
			// key2 was added.
			changes = append(changes, MapDiffChange{
				Key:      key2,
				NewValue: newMap[key2]})
			j++
		}
	}

	for ; i < len(oldKeys); i++ {
		// deleted key.
		changes = append(changes, MapDiffChange{
			Deleted: true,
			Key:     oldKeys[i]})
	}
	for ; j < len(newKeys); j++ {
		// added key.
		key := newKeys[j]
		changes = append(changes, MapDiffChange{
			Key:      key,
			NewValue: newMap[key]})
	}

	return changes
}

// MapSortedKeys returns the keys in the map in sorted order.
func mapSortedKeys(m map[string]string) []string {
	keys := make([]string, len(m))
	i := 0
	for k, _ := range m {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	return keys
}

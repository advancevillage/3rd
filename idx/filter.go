package idx

type Filter map[string]any

func Eq(field string, value any) Filter {
	return Filter{field: map[string]any{"$eq": value}}
}

func Gt(field string, value any) Filter {
	return Filter{field: map[string]any{"$gt": value}}
}

func In(field string, values ...any) Filter {
	return Filter{field: map[string]any{"$in": values}}
}

func And(filters ...Filter) Filter {
	items := make([]Filter, 0, len(filters))
	for _, filter := range filters {
		if len(filter) == 0 {
			continue
		}
		items = append(items, filter)
	}
	return Filter{"$and": items}
}

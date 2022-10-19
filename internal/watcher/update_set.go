package watcher

type UpdateSet struct {
	ToAdd    []DNSRecord
	ToUpdate []DNSRecord
	ToDelete []DNSRecord
}

func BuildUpdateSet(actual, expected []DNSRecord) UpdateSet {
	expectedRRs := make(map[string]DNSRecord, len(expected))
	for _, rr := range expected {
		expectedRRs[rr.Type+rr.Name] = rr
	}

	var out UpdateSet
	for _, rr := range actual {
		rrKey := rr.Type + rr.Name
		expectedRR, ok := expectedRRs[rrKey]
		if !ok {
			out.ToDelete = append(out.ToDelete, rr)
			continue
		}

		delete(expectedRRs, rrKey)
		if rr.IsEqual(expectedRR) {
			continue
		}

		rr.Content = expectedRR.Content
		rr.TTL = expectedRR.TTL
		rr.Proxied = expectedRR.Proxied
		out.ToUpdate = append(out.ToUpdate, rr)
	}

	for _, rr := range expectedRRs {
		out.ToAdd = append(out.ToAdd, rr)
	}

	return out
}

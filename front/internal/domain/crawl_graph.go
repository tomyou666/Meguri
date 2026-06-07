package domain

// ForwardReachableExisting はモード3用に選択ノードから有向到達可能な既存ノード ID を BFS 順で返す。
func ForwardReachableExisting(startID string, nodeIDs map[string]struct{}, outEdges map[string][]string) []string {
	visited := map[string]struct{}{startID: {}}
	order := []string{}
	queue := []string{startID}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		if current != startID {
			order = append(order, current)
		}
		for _, target := range outEdges[current] {
			if _, ok := nodeIDs[target]; !ok {
				continue
			}
			if _, seen := visited[target]; seen {
				continue
			}
			visited[target] = struct{}{}
			queue = append(queue, target)
		}
	}
	return order
}

// BuildOutEdges は source -> []target の隣接リストを構築する。
func BuildOutEdges(edges []struct{ Source, Target string }) map[string][]string {
	out := map[string][]string{}
	for _, e := range edges {
		out[e.Source] = append(out[e.Source], e.Target)
	}
	return out
}

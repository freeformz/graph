package graph

import (
	"testing"
)

func TestDirected_Traits(t *testing.T) {
	tests := map[string]struct {
		traits   *Traits
		expected *Traits
	}{
		"default traits": {
			traits:   &Traits{},
			expected: &Traits{},
		},
		"directed": {
			traits:   &Traits{IsDirected: true},
			expected: &Traits{IsDirected: true},
		},
	}

	for name, test := range tests {
		g := newDirected(IntHash, test.traits)
		traits := g.Traits()

		if !traitsAreEqual(traits, test.expected) {
			t.Errorf("%s: traits expectancy doesn't match: expected %v, got %v", name, test.expected, traits)
		}
	}
}

func TestDirected_Vertex(t *testing.T) {
	tests := map[string]struct {
		vertices         []int
		expectedVertices []int
	}{
		"graph with 3 vertices": {
			vertices:         []int{1, 2, 3},
			expectedVertices: []int{1, 2, 3},
		},
		"graph with duplicated vertex": {
			vertices:         []int{1, 2, 2},
			expectedVertices: []int{1, 2},
		},
	}

	for name, test := range tests {
		graph := newDirected(IntHash, &Traits{})

		for _, vertex := range test.vertices {
			graph.Vertex(vertex)
		}

		for _, vertex := range test.vertices {
			hash := graph.hash(vertex)
			if _, ok := graph.vertices[hash]; !ok {
				t.Errorf("%s: vertex %v not found in graph: %v", name, vertex, graph.vertices)
			}
		}
	}
}

func TestDirected_Edge(t *testing.T) {
	TestDirected_EdgeByHashes(t)
}

func TestDirected_EdgeByHashes(t *testing.T) {
	tests := map[string]struct {
		vertices      []int
		edges         []Edge[int]
		traits        *Traits
		expectedEdges []Edge[int]
		// Even though some of the EdgeByHashes calls might work, at least one of them could fail,
		// for example if the last call would introduce a cycle.
		shouldFinallyFail bool
	}{
		"graph with 2 edges": {
			vertices: []int{1, 2, 3},
			edges: []Edge[int]{
				{Source: 1, Target: 2, Properties: EdgeProperties{Weight: 10}},
				{Source: 1, Target: 3, Properties: EdgeProperties{Weight: 20}},
			},
			traits: &Traits{},
			expectedEdges: []Edge[int]{
				{Source: 1, Target: 2, Properties: EdgeProperties{Weight: 10}},
				{Source: 1, Target: 3, Properties: EdgeProperties{Weight: 20}},
			},
		},
		"hashes for non-existent vertices": {
			vertices: []int{1, 2},
			edges: []Edge[int]{
				{Source: 1, Target: 3, Properties: EdgeProperties{Weight: 20}},
			},
			traits:            &Traits{},
			shouldFinallyFail: true,
		},
		"edge introducing a cycle in an acyclic graph": {
			vertices: []int{1, 2, 3},
			edges: []Edge[int]{
				{Source: 1, Target: 2},
				{Source: 2, Target: 3},
				{Source: 3, Target: 1},
			},
			traits: &Traits{
				IsAcyclic: true,
			},
			shouldFinallyFail: true,
		},
		"edge with attributes": {
			vertices: []int{1, 2},
			edges: []Edge[int]{
				{
					Source: 1,
					Target: 2,
					Properties: EdgeProperties{
						Attributes: map[string]string{
							"color": "red",
						},
					},
				},
			},
			expectedEdges: []Edge[int]{
				{
					Source: 1,
					Target: 2,
					Properties: EdgeProperties{
						Attributes: map[string]string{
							"color": "red",
						},
					},
				},
			},
			traits: &Traits{},
		},
	}

	for name, test := range tests {
		graph := newDirected(IntHash, test.traits)

		for _, vertex := range test.vertices {
			graph.Vertex(vertex)
		}

		var err error

		for _, edge := range test.edges {
			if len(edge.Properties.Attributes) == 0 {
				err = graph.EdgeByHashes(edge.Source, edge.Target, EdgeWeight(edge.Properties.Weight))
			}
			// If there are edge attributes, iterate over them and call EdgeAttribute for each
			// entry. An edge should only have one attribute so that EdgeByHashes is invoked once.
			for key, value := range edge.Properties.Attributes {
				err = graph.EdgeByHashes(edge.Source, edge.Target, EdgeWeight(edge.Properties.Weight), EdgeAttribute(key, value))
			}
			if err != nil {
				break
			}
		}

		if test.shouldFinallyFail != (err != nil) {
			t.Fatalf("%s: error expectancy doesn't match: expected %v, got %v (error: %v)", name, test.shouldFinallyFail, (err != nil), err)
		}

		for _, expectedEdge := range test.expectedEdges {
			sourceHash := graph.hash(expectedEdge.Source)
			targetHash := graph.hash(expectedEdge.Target)

			edge, ok := graph.edges[sourceHash][targetHash]
			if !ok {
				t.Fatalf("%s: edge with source %v and target %v not found", name, expectedEdge.Source, expectedEdge.Target)
			}

			if edge.Source != expectedEdge.Source {
				t.Errorf("%s: edge sources don't match: expected source %v, got %v", name, expectedEdge.Source, edge.Source)
			}

			if edge.Target != expectedEdge.Target {
				t.Errorf("%s: edge targets don't match: expected target %v, got %v", name, expectedEdge.Target, edge.Target)
			}

			if edge.Properties.Weight != expectedEdge.Properties.Weight {
				t.Errorf("%s: edge weights don't match: expected weight %v, got %v", name, expectedEdge.Properties.Weight, edge.Properties.Weight)
			}

			if len(edge.Properties.Attributes) != len(expectedEdge.Properties.Attributes) {
				t.Fatalf("%s: attributes length don't match: expcted %v, got %v", name, len(expectedEdge.Properties.Attributes), len(edge.Properties.Attributes))
			}

			for expectedKey, expectedValue := range expectedEdge.Properties.Attributes {
				value, ok := edge.Properties.Attributes[expectedKey]
				if !ok {
					t.Errorf("%s: attribute keys don't match: expected key %v not found", name, expectedKey)
				}
				if value != expectedValue {
					t.Errorf("%s: attribute values don't match: expected value %v for key %v, got %v", name, expectedValue, expectedKey, value)
				}
			}
		}
	}
}

func TestDirected_GetEdge(t *testing.T) {
	TestDirected_GetEdgeByHashes(t)
}

func TestDirected_GetEdgeByHashes(t *testing.T) {
	tests := map[string]struct {
		vertices      []int
		getEdgeHashes [2]int
		exists        bool
	}{
		"get edge of directed graph": {
			vertices:      []int{1, 2, 3},
			getEdgeHashes: [2]int{1, 2},
			exists:        true,
		},
		"get non-existent edge of directed graph": {
			vertices:      []int{1, 2, 3},
			getEdgeHashes: [2]int{1, 3},
		},
	}

	for name, test := range tests {
		graph := newDirected(IntHash, &Traits{})
		for _, vertex := range test.vertices {
			graph.Vertex(vertex)
		}

		sourceHash := graph.hash(test.vertices[0])
		targetHash := graph.hash(test.vertices[1])

		graph.EdgeByHashes(sourceHash, targetHash)

		_, ok := graph.GetEdgeByHashes(test.getEdgeHashes[0], test.getEdgeHashes[1])

		if test.exists != ok {
			t.Fatalf("%s: result expectancy doesn't match: expected %v, got %v", name, test.exists, ok)
		}
	}
}

func TestDirected_DFS(t *testing.T) {
	TestDirected_DFSByHash(t)
}

func TestDirected_DFSByHash(t *testing.T) {
	tests := map[string]struct {
		vertices       []int
		edges          []Edge[int]
		startHash      int
		expectedVisits []int
		stopAtVertex   int
	}{
		"traverse entire graph with 3 vertices": {
			vertices: []int{1, 2, 3},
			edges: []Edge[int]{
				{Source: 1, Target: 2},
				{Source: 1, Target: 3},
			},
			startHash:      1,
			expectedVisits: []int{1, 2, 3},
			stopAtVertex:   -1,
		},
		"traverse entire triangle graph": {
			vertices: []int{1, 2, 3},
			edges: []Edge[int]{
				{Source: 1, Target: 2},
				{Source: 2, Target: 3},
				{Source: 3, Target: 1},
			},
			startHash:      1,
			expectedVisits: []int{1, 2, 3},
			stopAtVertex:   -1,
		},
		"traverse graph with 3 vertices until vertex 2": {
			vertices: []int{1, 2, 3},
			edges: []Edge[int]{
				{Source: 1, Target: 2},
				{Source: 2, Target: 3},
				{Source: 3, Target: 1},
			},
			startHash:      1,
			expectedVisits: []int{1, 2},
			stopAtVertex:   2,
		},
		"traverse a disconnected graph": {
			vertices: []int{1, 2, 3, 4},
			edges: []Edge[int]{
				{Source: 1, Target: 2},
				{Source: 3, Target: 4},
			},
			startHash:      1,
			expectedVisits: []int{1, 2},
			stopAtVertex:   -1,
		},
	}

	for name, test := range tests {
		graph := newDirected(IntHash, &Traits{})

		for _, vertex := range test.vertices {
			graph.Vertex(vertex)
		}

		for _, edge := range test.edges {
			if err := graph.Edge(edge.Source, edge.Target); err != nil {
				t.Fatalf("%s: failed to add edge: %s", name, err.Error())
			}
		}

		visited := make(map[int]struct{})

		visit := func(value int) bool {
			visited[value] = struct{}{}

			if test.stopAtVertex != -1 {
				if value == test.stopAtVertex {
					return true
				}
			}
			return false
		}

		_ = graph.DFSByHash(test.startHash, visit)

		if len(visited) != len(test.expectedVisits) {
			t.Fatalf("%s: numbers of visited vertices don't match: expected %v, got %v", name, len(test.expectedVisits), len(visited))
		}

		for _, expectedVisit := range test.expectedVisits {
			if _, ok := visited[expectedVisit]; !ok {
				t.Errorf("%s: expected vertex %v to be visited, but it isn't", name, expectedVisit)
			}
		}
	}
}

func TestDirected_BFS(t *testing.T) {
	TestDirected_BFSByHash(t)
}

func TestDirected_BFSByHash(t *testing.T) {
	tests := map[string]struct {
		vertices       []int
		edges          []Edge[int]
		startHash      int
		expectedVisits []int
		stopAtVertex   int
	}{
		"traverse entire graph with 3 vertices": {
			vertices: []int{1, 2, 3},
			edges: []Edge[int]{
				{Source: 1, Target: 2},
				{Source: 1, Target: 3},
			},
			startHash:      1,
			expectedVisits: []int{1, 2, 3},
			stopAtVertex:   -1,
		},
		"traverse graph with 6 vertices until vertex 4": {
			vertices: []int{1, 2, 3, 4, 5, 6},
			edges: []Edge[int]{
				{Source: 1, Target: 2},
				{Source: 1, Target: 3},
				{Source: 2, Target: 4},
				{Source: 2, Target: 5},
				{Source: 3, Target: 6},
			},
			startHash:      1,
			expectedVisits: []int{1, 2, 3, 4},
			stopAtVertex:   4,
		},
		"traverse a disconnected graph": {
			vertices: []int{1, 2, 3, 4},
			edges: []Edge[int]{
				{Source: 1, Target: 2},
				{Source: 3, Target: 4},
			},
			startHash:      1,
			expectedVisits: []int{1, 2},
			stopAtVertex:   -1,
		},
	}

	for name, test := range tests {
		graph := newDirected(IntHash, &Traits{})

		for _, vertex := range test.vertices {
			graph.Vertex(vertex)
		}

		for _, edge := range test.edges {
			if err := graph.Edge(edge.Source, edge.Target); err != nil {
				t.Fatalf("%s: failed to add edge: %s", name, err.Error())
			}
		}

		visited := make(map[int]struct{})

		visit := func(value int) bool {
			visited[value] = struct{}{}

			if test.stopAtVertex != -1 {
				if value == test.stopAtVertex {
					return true
				}
			}
			return false
		}

		_ = graph.BFSByHash(test.startHash, visit)

		for _, expectedVisit := range test.expectedVisits {
			if _, ok := visited[expectedVisit]; !ok {
				t.Errorf("%s: expected vertex %v to be visited, but it isn't", name, expectedVisit)
			}
		}
	}
}

func TestDirected_CreatesCycle(t *testing.T) {
	TestDirected_CreatesCycleByHashes(t)
}

func TestDirected_CreatesCycleByHashes(t *testing.T) {
	tests := map[string]struct {
		vertices     []int
		edges        []Edge[int]
		sourceHash   int
		targetHash   int
		createsCycle bool
	}{
		"directed 2-4-7-5 cycle": {
			vertices: []int{1, 2, 3, 4, 5, 6, 7},
			edges: []Edge[int]{
				{Source: 1, Target: 2},
				{Source: 1, Target: 3},
				{Source: 2, Target: 4},
				{Source: 3, Target: 6},
				{Source: 4, Target: 7},
				{Source: 5, Target: 2},
			},
			sourceHash:   7,
			targetHash:   5,
			createsCycle: true,
		},
		"undirected 2-4-7-5 'cycle'": {
			vertices: []int{1, 2, 3, 4, 5, 6, 7},
			edges: []Edge[int]{
				{Source: 1, Target: 2},
				{Source: 1, Target: 3},
				{Source: 2, Target: 4},
				{Source: 3, Target: 6},
				{Source: 4, Target: 7},
				{Source: 5, Target: 2},
			},
			sourceHash: 5,
			targetHash: 7,
			// The direction of the edge (57 instead of 75) doesn't create a directed cycle.
			createsCycle: false,
		},
		"no cycle": {
			vertices: []int{1, 2, 3, 4, 5, 6, 7},
			edges: []Edge[int]{
				{Source: 1, Target: 2},
				{Source: 1, Target: 3},
				{Source: 2, Target: 4},
				{Source: 3, Target: 6},
				{Source: 4, Target: 7},
				{Source: 5, Target: 2},
			},
			sourceHash:   5,
			targetHash:   6,
			createsCycle: false,
		},
	}

	for name, test := range tests {
		graph := newDirected(IntHash, &Traits{})

		for _, vertex := range test.vertices {
			graph.Vertex(vertex)
		}

		for _, edge := range test.edges {
			if err := graph.Edge(edge.Source, edge.Target); err != nil {
				t.Fatalf("%s: failed to add edge: %s", name, err.Error())
			}
		}

		createsCycle, err := graph.CreatesCycle(test.sourceHash, test.targetHash)
		if err != nil {
			t.Fatalf("%s: failed to add edge: %s", name, err.Error())
		}

		if createsCycle != test.createsCycle {
			t.Errorf("%s: cycle expectancy doesn't match: expected %v, got %v", name, test.createsCycle, createsCycle)
		}
	}
}

func TestDirected_Degree(t *testing.T) {
	TestDirected_DegreeByHash(t)
}

func TestDirected_DegreeByHash(t *testing.T) {
	tests := map[string]struct {
		vertices       []int
		edges          []Edge[int]
		vertex         int
		expectedDegree int
		shouldFail     bool
	}{
		"graph with 3 vertices and 2 edges": {
			vertices: []int{1, 2, 3},
			edges: []Edge[int]{
				{Source: 1, Target: 2},
				{Source: 1, Target: 3},
			},
			vertex:         1,
			expectedDegree: 2,
		},
		"graph with 3 vertices and 3 edges": {
			vertices: []int{1, 2, 3},
			edges: []Edge[int]{
				{Source: 1, Target: 2},
				{Source: 2, Target: 3},
				{Source: 3, Target: 1},
			},
			vertex:         2,
			expectedDegree: 2,
		},
		"disconnected graph": {
			vertices: []int{1, 2, 3},
			edges: []Edge[int]{
				{Source: 1, Target: 2},
			},
			vertex:         3,
			expectedDegree: 0,
		},
		"non-existent vertex": {
			vertices:   []int{1, 2, 3},
			vertex:     4,
			shouldFail: true,
		},
	}

	for name, test := range tests {
		graph := newDirected(IntHash, &Traits{})

		for _, vertex := range test.vertices {
			graph.Vertex(vertex)
		}

		for _, edge := range test.edges {
			if err := graph.Edge(edge.Source, edge.Target); err != nil {
				t.Fatalf("%s: failed to add edge: %s", name, err.Error())
			}
		}

		degree, err := graph.DegreeByHash(test.vertex)

		if test.shouldFail != (err != nil) {
			t.Fatalf("%s: error expectancy doesn't match: expected %v, got %v (error: %v)", name, test.shouldFail, (err != nil), err)
		}

		if degree != test.expectedDegree {
			t.Errorf("%s: degree expectancy doesn't match: expcted %v, got %v", name, test.expectedDegree, degree)
		}
	}
}

func TestDirected_StronglyConnectedComponents(t *testing.T) {
	tests := map[string]struct {
		vertices     []int
		edges        []Edge[int]
		expectedSCCs [][]int
	}{
		"graph with SCCs as on img/scc.svg": {
			vertices: []int{1, 2, 3, 4, 5, 6, 7, 8},
			edges: []Edge[int]{
				{Source: 1, Target: 2},
				{Source: 2, Target: 3},
				{Source: 2, Target: 5},
				{Source: 2, Target: 6},
				{Source: 3, Target: 4},
				{Source: 3, Target: 7},
				{Source: 4, Target: 3},
				{Source: 4, Target: 8},
				{Source: 5, Target: 1},
				{Source: 5, Target: 6},
				{Source: 6, Target: 7},
				{Source: 7, Target: 6},
				{Source: 8, Target: 4},
				{Source: 8, Target: 7},
			},
			expectedSCCs: [][]int{{1, 2, 5}, {3, 4, 8}, {6, 7}},
		},
	}

	for name, test := range tests {
		graph := newDirected(IntHash, &Traits{})

		for _, vertex := range test.vertices {
			graph.Vertex(vertex)
		}

		for _, edge := range test.edges {
			if err := graph.Edge(edge.Source, edge.Target); err != nil {
				t.Fatalf("%s: failed to add edge: %s", name, err.Error())
			}
		}

		sccs, _ := graph.StronglyConnectedComponents()
		matchedSCCs := 0

		for _, scc := range sccs {
			for _, expectedSCC := range test.expectedSCCs {
				if slicesAreEqual(scc, expectedSCC) {
					matchedSCCs++
				}
			}
		}

		if matchedSCCs != len(test.expectedSCCs) {
			t.Errorf("%s: expected SCCs don't match: expected %v, got %v", name, test.expectedSCCs, sccs)
		}
	}
}

func TestDirected_ShortesPath(t *testing.T) {
	TestDirected_ShortestPathByHashes(t)
}

func TestDirected_ShortestPathByHashes(t *testing.T) {
	tests := map[string]struct {
		vertices             []string
		edges                []Edge[string]
		sourceHash           string
		targetHash           string
		expectedShortestPath []string
		shouldFail           bool
	}{
		"graph as on img/dijkstra.svg": {
			vertices: []string{"A", "B", "C", "D", "E", "F", "G"},
			edges: []Edge[string]{
				{Source: "A", Target: "C", Properties: EdgeProperties{Weight: 3}},
				{Source: "A", Target: "F", Properties: EdgeProperties{Weight: 2}},
				{Source: "C", Target: "D", Properties: EdgeProperties{Weight: 4}},
				{Source: "C", Target: "E", Properties: EdgeProperties{Weight: 1}},
				{Source: "C", Target: "F", Properties: EdgeProperties{Weight: 2}},
				{Source: "D", Target: "B", Properties: EdgeProperties{Weight: 1}},
				{Source: "E", Target: "B", Properties: EdgeProperties{Weight: 2}},
				{Source: "E", Target: "F", Properties: EdgeProperties{Weight: 3}},
				{Source: "F", Target: "G", Properties: EdgeProperties{Weight: 5}},
				{Source: "G", Target: "B", Properties: EdgeProperties{Weight: 2}},
			},
			sourceHash:           "A",
			targetHash:           "B",
			expectedShortestPath: []string{"A", "C", "E", "B"},
		},
		"diamond-shaped graph": {
			vertices: []string{"A", "B", "C", "D"},
			edges: []Edge[string]{
				{Source: "A", Target: "B", Properties: EdgeProperties{Weight: 2}},
				{Source: "A", Target: "C", Properties: EdgeProperties{Weight: 4}},
				{Source: "B", Target: "D", Properties: EdgeProperties{Weight: 2}},
				{Source: "C", Target: "D", Properties: EdgeProperties{Weight: 2}},
			},
			sourceHash:           "A",
			targetHash:           "D",
			expectedShortestPath: []string{"A", "B", "D"},
		},
		"source equal to target": {
			vertices: []string{"A", "B", "C", "D"},
			edges: []Edge[string]{
				{Source: "A", Target: "B", Properties: EdgeProperties{Weight: 2}},
				{Source: "A", Target: "C", Properties: EdgeProperties{Weight: 4}},
				{Source: "B", Target: "D", Properties: EdgeProperties{Weight: 2}},
				{Source: "C", Target: "D", Properties: EdgeProperties{Weight: 2}},
			},
			sourceHash:           "B",
			targetHash:           "B",
			expectedShortestPath: []string{"B"},
		},
		"target not reachable": {
			vertices: []string{"A", "B", "C", "D"},
			edges: []Edge[string]{
				{Source: "A", Target: "B", Properties: EdgeProperties{Weight: 2}},
				{Source: "A", Target: "C", Properties: EdgeProperties{Weight: 4}},
			},
			sourceHash:           "A",
			targetHash:           "D",
			expectedShortestPath: []string{},
			shouldFail:           true,
		},
	}

	for name, test := range tests {
		graph := newDirected(StringHash, &Traits{})

		for _, vertex := range test.vertices {
			graph.Vertex(vertex)
		}

		for _, edge := range test.edges {
			if err := graph.Edge(edge.Source, edge.Target, EdgeWeight(edge.Properties.Weight)); err != nil {
				t.Fatalf("%s: failed to add edge: %s", name, err.Error())
			}
		}

		shortestPath, err := graph.ShortestPathByHashes(test.sourceHash, test.targetHash)

		if test.shouldFail != (err != nil) {
			t.Fatalf("%s: error expectancy doesn't match: expected %v, got %v (error: %v)", name, test.shouldFail, (err != nil), err)
		}

		if len(shortestPath) != len(test.expectedShortestPath) {
			t.Fatalf("%s: path length expectancy doesn't match: expected %v, got %v", name, len(test.expectedShortestPath), len(shortestPath))
		}

		for i, expectedVertex := range test.expectedShortestPath {
			if shortestPath[i] != expectedVertex {
				t.Errorf("%s: path vertex expectancy doesn't match: expected %v at index %d, got %v", name, expectedVertex, i, shortestPath[i])
			}
		}
	}
}

func TestDirected_AdjacencyList(t *testing.T) {
	tests := map[string]struct {
		vertices []int
		edges    []Edge[int]
		expected map[int]map[int]Edge[int]
	}{
		"Y-shaped graph": {
			vertices: []int{1, 2, 3, 4},
			edges: []Edge[int]{
				{Source: 1, Target: 3},
				{Source: 2, Target: 3},
				{Source: 3, Target: 4},
			},
			expected: map[int]map[int]Edge[int]{
				1: {
					3: {Source: 1, Target: 3},
				},
				2: {
					3: {Source: 2, Target: 3},
				},
				3: {
					4: {Source: 3, Target: 4},
				},
				4: {},
			},
		},
		"diamond-shaped graph": {
			vertices: []int{1, 2, 3, 4},
			edges: []Edge[int]{
				{Source: 1, Target: 2},
				{Source: 1, Target: 3},
				{Source: 2, Target: 4},
				{Source: 3, Target: 4},
			},
			expected: map[int]map[int]Edge[int]{
				1: {
					2: {Source: 1, Target: 2},
					3: {Source: 1, Target: 3},
				},
				2: {
					4: {Source: 2, Target: 4},
				},
				3: {
					4: {Source: 3, Target: 4},
				},
				4: {},
			},
		},
	}

	for name, test := range tests {
		graph := newDirected(IntHash, &Traits{})

		for _, vertex := range test.vertices {
			graph.Vertex(vertex)
		}

		for _, edge := range test.edges {
			if err := graph.Edge(edge.Source, edge.Target, EdgeWeight(edge.Properties.Weight)); err != nil {
				t.Fatalf("%s: failed to add edge: %s", name, err.Error())
			}
		}

		adjacencyMap := graph.AdjacencyMap()

		for expectedVertex, expectedAdjacencies := range test.expected {
			adjacencies, ok := adjacencyMap[expectedVertex]
			if !ok {
				t.Errorf("%s: expected vertex %v does not exist in adjacency map", name, expectedVertex)
			}

			for expectedAdjacency, expectedEdge := range expectedAdjacencies {
				edge, ok := adjacencies[expectedAdjacency]
				if !ok {
					t.Errorf("%s: expected adjacency %v does not exist in map of %v", name, expectedAdjacency, expectedVertex)
				}
				if edge.Source != expectedEdge.Source || edge.Target != expectedEdge.Target {
					t.Errorf("%s: edge expectancy doesn't match: expected %v, got %v", name, expectedEdge, edge)
				}
			}
		}
	}
}

func TestDirected_edgesAreEqual(t *testing.T) {
	tests := map[string]struct {
		a             Edge[int]
		b             Edge[int]
		edgesAreEqual bool
	}{
		"equal edges in directed graph": {
			a:             Edge[int]{Source: 1, Target: 2},
			b:             Edge[int]{Source: 1, Target: 2},
			edgesAreEqual: true,
		},
		"swapped equal edges in directed graph": {
			a: Edge[int]{Source: 1, Target: 2},
			b: Edge[int]{Source: 2, Target: 1},
		},
	}

	for name, test := range tests {
		graph := newDirected(IntHash, &Traits{})
		actual := graph.edgesAreEqual(test.a, test.b)

		if actual != test.edgesAreEqual {
			t.Errorf("%s: equality expectations don't match: expected %v, got %v", name, test.edgesAreEqual, actual)
		}
	}
}

func TestDirected_addEdge(t *testing.T) {
	tests := map[string]struct {
		edges []Edge[int]
	}{
		"add 3 edges": {
			edges: []Edge[int]{
				{Source: 1, Target: 2, Properties: EdgeProperties{Weight: 1}},
				{Source: 2, Target: 3, Properties: EdgeProperties{Weight: 2}},
				{Source: 3, Target: 1, Properties: EdgeProperties{Weight: 3}},
			},
		},
	}

	for name, test := range tests {
		graph := newDirected(IntHash, &Traits{})

		for _, edge := range test.edges {
			sourceHash := graph.hash(edge.Source)
			TargetHash := graph.hash(edge.Target)
			graph.addEdge(sourceHash, TargetHash, edge)
		}

		if len(graph.edges) != len(test.edges) {
			t.Errorf("%s: number of edges doesn't match: expected %v, got %v", name, len(test.edges), len(graph.edges))
		}
		if len(graph.outEdges) != len(test.edges) {
			t.Errorf("%s: number of outgoing edges doesn't match: expected %v, got %v", name, len(test.edges), len(graph.outEdges))
		}
		if len(graph.inEdges) != len(test.edges) {
			t.Errorf("%s: number of ingoing edges doesn't match: expected %v, got %v", name, len(test.edges), len(graph.inEdges))
		}
	}
}

func TestDirected_predecessors(t *testing.T) {
	tests := map[string]struct {
		vertices             []int
		edges                []Edge[int]
		vertex               int
		expectedPredecessors []int
	}{
		"graph with 3 vertices": {
			vertices: []int{1, 2, 3},
			edges: []Edge[int]{
				{Source: 1, Target: 2},
				{Source: 1, Target: 3},
			},
			vertex:               2,
			expectedPredecessors: []int{1},
		},
		"graph with 6 vertices": {
			vertices: []int{1, 2, 3, 4, 5, 6},
			edges: []Edge[int]{
				{Source: 1, Target: 2},
				{Source: 1, Target: 3},
				{Source: 2, Target: 4},
				{Source: 2, Target: 5},
				{Source: 3, Target: 6},
			},
			vertex:               5,
			expectedPredecessors: []int{2},
		},
		"graph with 4 vertices and 3 predecessors": {
			vertices: []int{1, 2, 3, 4},
			edges: []Edge[int]{
				{Source: 1, Target: 4},
				{Source: 2, Target: 4},
				{Source: 3, Target: 4},
			},
			vertex:               4,
			expectedPredecessors: []int{1, 2, 3},
		},
	}

	for name, test := range tests {
		graph := newDirected(IntHash, &Traits{})

		for _, vertex := range test.vertices {
			graph.Vertex(vertex)
		}

		for _, edge := range test.edges {
			if err := graph.Edge(edge.Source, edge.Target); err != nil {
				t.Fatalf("%s: failed to add edge: %s", name, err.Error())
			}
		}

		predecessors := graph.predecessors(graph.hash(test.vertex))

		if !slicesAreEqual(predecessors, test.expectedPredecessors) {
			t.Errorf("%s: predecessors don't match: expected %v, got %v", name, test.expectedPredecessors, predecessors)
		}
	}
}

func slicesAreEqual[T comparable](a []T, b []T) bool {
	if len(a) != len(b) {
		return false
	}

	for _, aValue := range a {
		found := false
		for _, bValue := range b {
			if aValue == bValue {
				found = true
			}
		}
		if !found {
			return false
		}
	}

	return true
}

package LogootOpBased

import (
	"fmt"
	"math"
	"math/rand"

	"strconv"
)

type Identifier struct {
	site     string
	position int // position must be an integer between 0 and 999
}

type Position []Identifier

func (pos Position) Tostring() string {
	s := "{\n"

	for x := range pos {
		s += pos[x].site + ", " + strconv.Itoa(pos[x].position) + "\n"
	}
	s += "}"
	return s
}

func (pos Position) Add(x int) Position {
	list := make(Position, 0)

	for pos_elem := range pos {
		list = append(list, pos[pos_elem])
	}

	n := len(pos) - 1
	to_add := x
	for to_add > 0 {
		space := MaxPos - list[n].position

		if space < to_add {
			index := n
			for index > 0 {
				if list[index].position < MaxPos {
					break
				} else {
					list[index].position = 0
					index = index - 1
				}
			}
			list[index].position = list[index].position + 1
			to_add = to_add - space
		} else {
			list[n].position = list[n].position + to_add
			to_add = 0
		}

	}

	return list
}

type PositionIdentifier struct {
	clock int
	pos   Position
}

var MaxPos = 999

func (p Identifier) LowerThan(q Identifier) bool {
	if p.position < q.position {
		return true
	} else if q.position == p.position && p.site < q.site {
		return true
	} else {
		return false
	}
}

func (p Position) LowerThan(q Position) bool {
	b := true
	i := 0
	for i < len(p) {
		if q[i].LowerThan(p[i]) {
			b = false
			break
		} else if p[i].LowerThan(q[i]) {
			break
		}
		i = i + 1
	}
	return b
}

func (pos Position) Prefix(index int) Position {
	list := make(Position, 0)
	i := 0
	if index > 0 {
		for pos_elem := range pos {
			list = append(list, pos[pos_elem])
			i = i + 1
			if i >= index {
				break
			}
		}
		if i < index {
			for i < index {
				list = append(list, Identifier{site: "", position: 0})
				i = i + 1
			}
		}
	}
	return list
}

func Distance(p Position, q Position) uint64 {
	if p.LowerThan(q) {

		i := 0
		index := 0
		for i < len(p) {
			if p[i].position == q[i].position {
				i = i + 1
			} else {
				index = i
				break
			}
		}

		var distance (uint64) = 0
		n := int(math.Max(float64(len(p)), float64(len(q))))
		i = index
		distance = (uint64)(q[i].position - p[i].position)
		for i < n {
			pmin := -1
			qmax := MaxPos + 1
			if i < len(p) {
				pmin = p[i].position
			}
			if i < len(q) {
				qmax = q[i].position
			}
			distance = distance * (uint64)(MaxPos-pmin+qmax)
			i = i + 1
		}

		return distance
	} else if !q.LowerThan(p) {
		return 0

	} else {
		return Distance(q, p)
	}
}

func FirstPosition(id string) Position {
	list := make(Position, 1)
	list[0] = Identifier{
		site: id, position: 0}
	return list
}

func LastPosition(id string) Position {
	list := make(Position, 1)
	list[0] = Identifier{
		site: id, position: MaxPos}
	return list
}

func constructPosition(r Position, p Position, q Position, s string) Position {
	list := make([]Identifier, 0)
	n := len(r)
	for i, pos_elem := range r {
		if i == n {
			list = append(list, Identifier{site: s, position: pos_elem.position})
		} else if p[i].position == pos_elem.position {
			list = append(list, Identifier{site: p[i].site, position: pos_elem.position})
		} else if q[i].position == pos_elem.position {
			list = append(list, Identifier{site: q[i].site, position: pos_elem.position})
		} else {
			list = append(list, Identifier{site: s, position: pos_elem.position})
		}
	}
	return list

}

func generateLinePositions(p Position, q Position, N int, s string) []Position {
	if p.LowerThan(q) {
		fmt.Printf("Hello 1 \n")
		list := make([]Position, 0)
		index := 0
		var interval (uint64) = 0
		fmt.Printf("Hello 2 \n")

		for interval < uint64(N) {
			fmt.Printf("distance : %d \n", interval)
			index = index + 1
			interval = Distance(q.Prefix(index), p.Prefix(index))
		}
		fmt.Printf("Hello 3 \n")
		var Nlong uint64 = uint64(N)
		step := interval / Nlong
		r := p.Prefix(index)
		j := 1
		fmt.Printf("Hello 4 \n")
		for j <= N {
			list = append(list, constructPosition(r.Add(rand.Intn(int(step))+1), p, q, s))
			r = r.Add(int(step))
			j = j + 1
		}
		fmt.Printf("Hello 5 \n")
		return list
	} else if q.LowerThan(p) {
		return generateLinePositions(q, p, N, s)
	} else {
		panic(fmt.Errorf("cannot generate %d position between p: %s,  and q: %s. They are the same\n ", N, p.Tostring(), q.Tostring()))
	}
}

type LogootLine struct {
	Pos  PositionIdentifier
	Text string
}

// A Logoot text is a tree, lets define its node, that can contain sons, that have an Identifier (n, site) representing it and line != nil if this line exists (line corresponding since the begining)
type LogootTextNode struct {
	Sons  []*LogootTextNode
	PosNb Identifier
	Line  *LogootLine
}

type LogootText struct {
	Root *LogootTextNode
}

func (lgtxt LogootText) GetFirstNode() *LogootTextNode {
	b := true
	node := lgtxt.Root
	for b {
		if node.Line != nil {
			b = false
		} else {
			node = node.Sons[0]
		}
	}
	return node
}

func (lgtxt LogootText) GetLastNode() *LogootTextNode {
	node := lgtxt.Root
	for len(node.Sons) > 0 {
		node = node.Sons[len(node.Sons)-1]
	}
	return node
}

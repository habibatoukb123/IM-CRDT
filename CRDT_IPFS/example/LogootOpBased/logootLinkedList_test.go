package LogootOpBased

import (
	"testing"
)

// TestCreation calls FirstPosition and Last Postition.
// for a valid return value.
func TestCreation(t *testing.T) {
	site := "Gladys"
	PositionZero := FirstPosition(site)
	if PositionZero[0].position != 0 || PositionZero[0].site != site {
		t.Fatalf(`PositionZero("Gladys") = {[%s, %d}], But we wanted PositionZero("Gladys") = {["Gladys", 0}]`, PositionZero[0].site, PositionZero[0].position, MaxPos)
	}

	PositionEnd := LastPosition(site)
	if PositionEnd[0].position != 0 || PositionEnd[0].site != site {
		t.Fatalf(`LastPosition("Gladys") = {[%s, %d}], But we wanted LastPosition("Gladys") = {["Gladys", %d}]`, PositionZero[0].site, PositionZero[0].position, MaxPos)
	}
}

// TestPrefix calls Prefix and check its working way,
// checking for a valid return value.
func TestPrefix(t *testing.T) {
	site := "Gladys"
	pos := make(Position, 5)

	pos[0] = Identifier{site: site, position: 2}
	pos[1] = Identifier{site: site, position: 5}
	pos[2] = Identifier{site: site, position: 4}
	pos[3] = Identifier{site: site, position: 3}
	pos[4] = Identifier{site: site, position: 7}

	// Pos = (2,5,4,3,7)

	posBis := make(Position, 5)

	posBis[0] = Identifier{site: site, position: 2}
	posBis[1] = Identifier{site: site, position: 5}
	posBis[2] = Identifier{site: site, position: 4}

	// PosBis = (2,5,4)

	if !pos.Prefix(3).LowerThan(posBis) && !posBis.LowerThan(pos.Prefix(3)) {
		t.Fatalf(`(2,5,4,3,7).Prefix(3) != (2,5,4), instead we have : \n%s\n`, pos.Prefix(3).Tostring())
	}

	if len(pos.Prefix(0)) > 0 {
		t.Fatalf(`len((2,5,4,3,7).Prefix(0)) > 0 ,  we have : \n%d\n`, len(pos.Prefix(0)))
	}

}

func TestDistance(t *testing.T) {
	site := "Gladys"
	pos := make(Position, 5)

	pos[0] = Identifier{site: site, position: 2}
	pos[1] = Identifier{site: site, position: 5}
	pos[2] = Identifier{site: site, position: 4}
	pos[3] = Identifier{site: site, position: 3}
	pos[4] = Identifier{site: site, position: 7}

	// Pos = (2,5,4,3,7)

	posBis := make(Position, 5)

	posBis[0] = Identifier{site: site, position: 2}
	posBis[1] = Identifier{site: site, position: 5}
	posBis[2] = Identifier{site: site, position: 4}
	posBis[3] = Identifier{site: site, position: 3}
	posBis[4] = Identifier{site: site, position: 5}

	// PosBis = (2,5,4,3,7)

	if Distance(pos, posBis) != 2 {
		t.Fatalf(`Distance(pos, posBis) != 2 , instead we have : \n%d\n`, Distance(pos, posBis))
	}

	posBis[0] = Identifier{site: site, position: 2}
	posBis[1] = Identifier{site: site, position: 5}
	posBis[2] = Identifier{site: site, position: 4}
	posBis[3] = Identifier{site: site, position: 5}
	posBis[4] = Identifier{site: site, position: 4}
	// PosBis = (2,5,4,5,4)

	// distance((2,5,4,3,7),
	//          (2,5,4,5,4))
	if Distance(pos, posBis) != uint64((MaxPos - 7 + MaxPos + 4)) {
		t.Fatalf(`distance((2,5,4,3,7), (2,5,4,5,4)) = %d but we wanted : %d\n`, Distance(pos, posBis), MaxPos-7+MaxPos+4)
	}

	posBis[0] = Identifier{site: site, position: 2}
	posBis[1] = Identifier{site: site, position: 5}
	posBis[2] = Identifier{site: site, position: 4}
	posBis[3] = Identifier{site: site, position: 6}
	posBis[4] = Identifier{site: site, position: 5}
	// PosBis = (2,5,4,6,5)

	// distance((2,5,4,3,7),
	//          (2,5,4,6))
	if Distance(pos, posBis.Prefix(4)) != uint64((MaxPos - 7 + MaxPos + MaxPos)) {
		t.Fatalf(`distance((2,5,4,3,7), (2,5,4,5,4)) = %d but we wanted : %d\n`, Distance(pos, posBis), MaxPos-7+MaxPos+4)
	}

	if Distance(pos, pos) != 0 {
		t.Fatalf(`distance((2,5,4,3,7),(2,5,4,3,7)) = %d but we wanted : %d\n`, Distance(pos, posBis), 0)
	}

}

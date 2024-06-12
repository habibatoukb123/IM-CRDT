package LogootOpBased

import (
	CRDTDag "IPFS_CRDT/CRDTDag"
	"IPFS_CRDT/Config"
	CRDT "IPFS_CRDT/Crdt"
	Payload "IPFS_CRDT/Payload"
	IpfsLink "IPFS_CRDT/ipfsLink"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ipfs/go-cid"
	"github.com/ipfs/interface-go-ipfs-core/path"
	"golang.org/x/sync/semaphore"
)

// =======================================================================================
// Payload - OpBased
// =======================================================================================

type Element LogootLine
type OpNature int

const (
	INSERT OpNature = iota
	DELETE
)

type Operation struct {
	Elem Element
	Op   OpNature
}

func (self Operation) ToString() string {
	b, err := json.Marshal(self)
	if err != nil {
		panic(fmt.Errorf("Set Operation To string fail to Marshal\nError: %s", err))
	}
	return string(b[:])
}
func (op *Operation) op_from_string(s string) {
	err := json.Unmarshal([]byte(s), op)
	if err != nil {
		panic(fmt.Errorf("Set Operation To string fail to Marshal\nError: %s", err))
	}
}

type PayloadOpBased struct {
	Op Operation
	Id string
}

func (self *PayloadOpBased) Create_PayloadOpBased(s string, o1 Operation) {

	self.Op = o1
	self.Id = s
}
func (self *PayloadOpBased) ToString() string {
	b, err := json.Marshal(self)
	if err != nil {
		panic(fmt.Errorf("Set Operation To string fail to Marshal\nError: %s", err))
	}
	return string(b[:])
}
func (self *PayloadOpBased) FromString(s string) {
	err := json.Unmarshal([]byte(s), self)
	if err != nil {
		panic(fmt.Errorf("Set Operation To string fail to Marshal\nError: %s", err))
	}
}

// =======================================================================================
// CRDTSet OpBased
// =======================================================================================

// ===============Mark Ta Page=================
type CRDTLogootOpBased struct {
	sys *IpfsLink.IpfsLink
	txt LogootText
}

func Create_CRDTLogootOpBased(s *IpfsLink.IpfsLink) CRDTLogootOpBased {
	pi1 := LogootLine{PositionIdentifier{clock: 0, pos: FirstPosition("Creation")}, ""}
	pi2 := LogootLine{PositionIdentifier{clock: 0, pos: LastPosition("Creation")}, ""}

	node2 := LogootTextNode{Line: &pi2, PosNb: pi2.Pos.pos[0], Sons: make([]*LogootTextNode, 0)}
	so := make([]*LogootTextNode, 0)
	so = append(so, &node2)
	node1 := LogootTextNode{Line: &pi1, PosNb: pi1.Pos.pos[0], Sons: so}

	logoottxt := LogootText{Root: &node1}

	return CRDTLogootOpBased{
		sys: s,
		txt: logoottxt,
	}

}
func search(list []string, x string) int {
	for i := 0; i < len(list); i++ {
		if list[i] == x {
			return i
		}
	}
	return -1
}

func idEqual(id1 Identifier, id2 Identifier) bool {
	if id1.position == id2.position && id1.site == id2.site {
		return true
	}
	return false
}

// check if id1 < id2
func LowerThan(id1 Identifier, id2 Identifier) bool {
	if id1.position <= id2.position || (id1.position == id2.position && id1.site <= id2.site) {
		return true
	}
	return false
}

func (self *CRDTLogootOpBased) Insert(pid PositionIdentifier, text string) {

	node := self.txt.Root
	x := 0
	for x < len(pid.pos) {
		i := 0
		found := false
		for i < len(node.Sons) {
			// equality condition ( i.e. neitherpid.pos[x] < node.sons[i].posNb and neither node.sons[i].posNb < pid.pos[x] )
			if idEqual(node.Sons[i].PosNb, pid.pos[x]) {
				node = node.Sons[i]
				found = true
				break
			} else if LowerThan(node.Sons[i].PosNb, pid.pos[x]) {
				break
			} else {
				i = i + 1
			}
		}

		if !found {
			newNode := LogootTextNode{Sons: make([]*LogootTextNode, 0), PosNb: pid.pos[x], Line: nil}
			// if we create a node that is the exact position we were searching for, we must add the line to it
			if x >= len(pid.pos)-1 {
				line := LogootLine{Pos: pid, Text: text}
				newNode.Line = &line
			}
			if i >= len(node.Sons)-1 {
				node.Sons = append(node.Sons, &newNode)
			} else {
				node.Sons = append(node.Sons[:i+1], node.Sons[i:]...)
				node.Sons[i] = &newNode
			}
		} else if found && x == len(pid.pos)-1 {
			// if node is found and this is the last we are looking for, we must just add the line to this node.
			line := LogootLine{Pos: pid, Text: text}
			node.Line = &line

		}
		x = x + 1
	}

}

func (self *CRDTLogootOpBased) Delete(pid PositionIdentifier) {
	nodesIndex := make([]int, 0)
	path := make([]*LogootTextNode, 0)
	node := self.txt.Root
	path = append(path, node)
	nodesIndex = append(nodesIndex, -1)
	x := 0
	for x <= len(pid.pos) {
		i := 0
		found := false
		for i <= len(node.Sons) {
			// equality condition ( i.e. neitherpid.pos[x] < node.sons[i].posNb and neither node.sons[i].posNb < pid.pos[x] )
			if idEqual(node.Sons[i].PosNb, pid.pos[x]) {
				nodesIndex = append(nodesIndex, i)
				node = node.Sons[i]
				found = true
				path = append(path, node)
				break
			} else if LowerThan(node.Sons[i].PosNb, pid.pos[x]) {
				break
			} else {
				i = i + 1
			}
		}

		if !found {
			panic(fmt.Errorf("Error in Order of receival of updates I guess\n I received Deletion of %s, but I don't know it", pid.pos.Tostring()))
		} else if found && x == len(pid.pos)-1 {

			node.Line = nil
			// node isn't containing a line anymore, this is a empty position
			// if it isn't part of a trajectory of anyone, we should delete it (and hope golang compilator detect such unused memory as we cannot free it as in C)
			index := len(path) - 1
			for index > 0 {
				node := path[index]
				lastnodeIndex := nodesIndex[index-1]
				lastNode := path[index-1]
				if node.Line == nil && len(node.Sons) == 0 {
					if lastnodeIndex == 0 {
						lastNode.Sons = lastNode.Sons[1:]
					} else if lastnodeIndex == len(lastNode.Sons)-1 {
						lastNode.Sons = lastNode.Sons[:len(lastNode.Sons)-2]
					} else {
						lastNode.Sons = append(lastNode.Sons[:lastnodeIndex], lastNode.Sons[lastnodeIndex+1:]...)
					}
				} else {
					break
				}
				index = index - 1
			}
		}
		x = x + 1

	}
}

type PairPosString struct {
	P LogootLine
	S string
}

func lookupRecursif(node *LogootTextNode) []PairPosString {
	if node == nil {
		return make([]PairPosString, 0)
	} else {
		var list []PairPosString
		list = make([]PairPosString, 0)
		if node.Line != nil {
			list = append(list, PairPosString{P: *node.Line, S: node.Line.Text})
		}
		for x := range node.Sons {
			list = append(list, lookupRecursif(node.Sons[x])...)
		}

		return list
	}
}

func (self *CRDTLogootOpBased) Lookup() []PairPosString {
	return lookupRecursif(self.txt.Root)
}

func (self *CRDTLogootOpBased) ToFile(file string) {

	b, err := json.Marshal(self)
	if err != nil {
		panic(fmt.Errorf("CRDTDagNode - ToFile Could not Marshall %s\nError: %s", file, err))
	}
	f, err := os.Create(file)
	if err != nil {
		panic(fmt.Errorf("CRDTDagNode - ToFile Could not Create the file %s\nError: %s", file, err))
	}
	f.Write(b)
	err = f.Close()
	if err != nil {
		panic(fmt.Errorf("CRDTDagNode - ToFile Could not Write to the file %s\nError: %s", file, err))
	}
}

// =======================================================================================
// CRDTSetDagNode OpBased
// =======================================================================================

type CRDTLogootOpBasedDagNode struct {
	DagNode CRDTDag.CRDTDagNode
}

func (self *CRDTLogootOpBasedDagNode) FromFile(fil string) {
	var pl Payload.Payload = &PayloadOpBased{}
	self.DagNode.CreateNodeFromFile(fil, &pl)
}

func (self *CRDTLogootOpBasedDagNode) GetDirect_dependency() []CRDTDag.EncodedStr {

	return self.DagNode.DirectDependency
}

func (self *CRDTLogootOpBasedDagNode) ToFile(file string) {

	self.DagNode.ToFile(file)
}
func (self *CRDTLogootOpBasedDagNode) GetEvent() *Payload.Payload {

	return self.DagNode.Event
}
func (self *CRDTLogootOpBasedDagNode) GetPiD() string {

	return self.DagNode.PID
}
func (self *CRDTLogootOpBasedDagNode) CreateEmptyNode() *CRDTDag.CRDTDagNodeInterface {
	n := CreateDagNode(Operation{}, "")
	var node CRDTDag.CRDTDagNodeInterface = &n
	return &node
}
func CreateDagNode(o Operation, id string) CRDTLogootOpBasedDagNode {
	var pl Payload.Payload = &PayloadOpBased{Op: o, Id: id}
	slic := make([]CRDTDag.EncodedStr, 0)
	return CRDTLogootOpBasedDagNode{
		DagNode: CRDTDag.CRDTDagNode{
			Event:            &pl,
			PID:              id,
			DirectDependency: slic,
		},
	}
}

// =======================================================================================
// CRDTSetDag OpBased
// =======================================================================================

type CRDTLogootOpBasedDag struct {
	dag         CRDTDag.CRDTManager
	measurement bool
	Data        CRDTLogootOpBased
}

func (self *CRDTLogootOpBasedDag) GetDag() *CRDTDag.CRDTManager {

	return &self.dag
}
func (self *CRDTLogootOpBasedDag) SendRemoteUpdates() {

	self.dag.SendRemoteUpdates()
}
func (self *CRDTLogootOpBasedDag) GetCRDTManager() *CRDTDag.CRDTManager {

	return &self.dag
}
func (self *CRDTLogootOpBasedDag) IsKnown(cid CRDTDag.EncodedStr) bool {

	find := false
	for x := range self.dag.GetAllNodes() {
		if string(self.dag.GetAllNodes()[x]) == string(cid.Str) {
			find = true
			break
		}
	}
	return find
}
func (self *CRDTLogootOpBasedDag) Merge(cids []CRDTDag.EncodedStr) []string {

	to_add := make([]CRDTDag.EncodedStr, 0)
	for _, cid := range cids {
		find := self.IsKnown(cid)
		if !find {
			to_add = append(to_add, cid)
		}
	}

	fils, err := self.dag.GetNodeFromEncodedCid(to_add)
	if err != nil {
		panic(fmt.Errorf("could not get ndoes from encoded cids\nerror :%s", err))
	}

	for index := range fils {
		fil := fils[index]
		n := CreateDagNode(Operation{}, "") // Create an Empty operation
		n.FromFile(fil)                     // Fill it with the operation just read
		self.remoteAddNode(cids[index], n)  // Add the data as a Remote operation (which are applied as a local one)
	}

	self.Data = self.Lookup()
	return fils
}

func (self *CRDTLogootOpBasedDag) remoteAddNode(cID CRDTDag.EncodedStr, newnode CRDTLogootOpBasedDagNode) {
	var pl CRDTDag.CRDTDagNodeInterface = &newnode
	self.dag.RemoteAddNodeSuper(cID, &pl)
	self.Data = self.Lookup()
}

func (self *CRDTLogootOpBasedDag) callAddToIPFS(bytes []byte, file string) (path.Resolved, error) {
	time_toencrypt := -1
	ti := time.Now()
	var path path.Resolved
	var err error
	if self.dag.Key != "" {
		path, err = self.GetCRDTManager().AddToIPFS(self.dag.Sys, bytes, &time_toencrypt)

	} else {
		path, err = self.GetCRDTManager().AddToIPFS(self.dag.Sys, bytes)
		time_toencrypt = 0

	}
	if err != nil {
		panic(fmt.Errorf("Error in callAddToIPFS, Couldn't add file to IPFS\nError: %s\n", err))
	}
	Total_AddTime := int(time.Since(ti).Nanoseconds())
	time_add := Total_AddTime - time_toencrypt

	if self.measurement {
		// Write time to encrypt in a file
		fstrBis := ""
		if self.dag.Key != "" {
			fstrBis = file + ".timeEncrypt"
			if _, err := os.Stat(fstrBis); !errors.Is(err, os.ErrNotExist) {
				os.Remove(fstrBis)
			}
			fil, err := os.OpenFile(fstrBis, os.O_CREATE|os.O_WRONLY, 0755)
			if err != nil {
				panic(fmt.Errorf("Error RemoteAddNodeSupde - , Could not open the time file to write encoded data\nError: %s", err))
			}
			_, err = fil.Write([]byte(strconv.Itoa(time_toencrypt)))
			if err != nil {
				panic(fmt.Errorf("Error RemoteAddNodeSupde - , Could not write the time file to write encoded data\nError: %s", err))
			}
			err = fil.Close()
			if err != nil {
				panic(fmt.Errorf("Error RemoteAddNodeSupde - , Could not close the time file to write encoded data \nError: %s", err))
			}
		}

		// Write time to add to IFPS
		fstrBis = file + ".timeAdd"
		if _, err := os.Stat(fstrBis); !errors.Is(err, os.ErrNotExist) {
			os.Remove(fstrBis)
		}
		fil, err := os.OpenFile(fstrBis, os.O_CREATE|os.O_WRONLY, 0755)
		if err != nil {
			panic(fmt.Errorf("Error RemoteAddNodeSupde - , Could not open the time file to write encoded data\nError: %s", err))
		}
		_, err = fil.Write([]byte(strconv.Itoa(time_add)))
		if err != nil {
			panic(fmt.Errorf("Error RemoteAddNodeSupde - , Could not write the time file to write encoded data\nError: %s", err))
		}
		err = fil.Close()
		if err != nil {
			panic(fmt.Errorf("Error RemoteAddNodeSupde - , Could not close the time file to write encoded data \nError: %s", err))
		}
	}

	return path, err
}

func (self *CRDTLogootOpBasedDag) Insert(previousLine LogootLine, nextLine LogootLine, x string) (string, TimeTuple) {
	generatedPosition := generateLinePositions(previousLine.Pos.pos, nextLine.Pos.pos, 1, self.GetSys().IpfsNode.Identity.Pretty())
	op := Operation{Elem: Element(LogootLine{Pos: PositionIdentifier{clock: time.Now().Nanosecond(), pos: generatedPosition[0]}, Text: x}), Op: INSERT}
	newNode := CreateDagNode(op, self.GetSys().IpfsNode.Identity.Pretty())
	for dependency := range self.dag.Root_nodes {
		// fmt.Println("dep:", self.dag.Root_nodes[dependency].Str)
		newNode.DagNode.DirectDependency = append(newNode.DagNode.DirectDependency, self.dag.Root_nodes[dependency])
	}

	strFile := self.dag.NextFileName()
	if _, err := os.Stat(strFile); !errors.Is(err, os.ErrNotExist) {
		os.Remove(strFile)
	}
	newNode.ToFile(strFile)
	bytes, err := os.ReadFile(strFile)
	if err != nil {
		panic(fmt.Errorf("ERROR INCREMENT CRDTSetOpBasedDag, could not read file\nerror: %s", err))
	}
	path, err := self.callAddToIPFS(bytes, strFile)
	if err != nil {
		panic(fmt.Errorf("CRDTSetOpBasedDag Increment, could not add the file to IFPS\nerror: %s", err))
	}

	encodedCid := self.dag.EncodeCid(path)
	c := cid.Cid{}
	err = json.Unmarshal(encodedCid.Str, &c)
	if err != nil {
		panic(fmt.Errorf("CRDTSetOpBasedDag Increment, could not UnMarshal\nerror: %s", err))
	}

	// fmt.Println("encodedCid Increment :", c.String())
	var pl CRDTDag.CRDTDagNodeInterface = &newNode

	self.dag.AddNode(encodedCid, &pl) // Adding the node created before to the Merkle-DAG

	self.SendRemoteUpdates() // Op-Based force us to send updates to other at each update

	times := TimeTuple{} // Time measurement structure, for analysis only (when self.Measurement is true)

	if self.measurement {
		//Add time
		b, err := os.ReadFile(strFile + ".timeAdd")
		if err != nil {
			panic(fmt.Errorf("Couldn't read TimeAdd file\nError: %s\n", err))
		}
		intAdd, err := strconv.Atoi(string(b))
		if err != nil {
			panic(fmt.Errorf(" TimeAdd file is malformatted, and couldn't be Atoi'ed\nError: %s\n", err))
		}
		times.Time_add = intAdd

		err = os.Remove(strFile + ".timeAdd")
		if err != nil {
			panic(fmt.Errorf("Couldn't Remove TimeAdd file\nError: %s\n", err))
		}

		// Encrypt Time
		times.Time_encrypt = 0
		if self.dag.Key != "" {
			b, err = os.ReadFile(strFile + ".timeEncrypt")
			if err != nil {
				panic(fmt.Errorf("Couldn't read timeEncrypt file\nError: %s\n", err))
			}
			intAdd, err = strconv.Atoi(string(b))
			if err != nil {
				panic(fmt.Errorf(" timeEncrypt file is malformatted, and couldn't be Atoi'ed\nError: %s\n", err))
			}
			times.Time_encrypt = intAdd

			err = os.Remove(strFile + ".timeEncrypt")
			if err != nil {
				panic(fmt.Errorf("Couldn't Remove timeEncrypt file\nError: %s\n", err))
			}
		}

	}
	self.Data = self.Lookup()

	return c.String(), times
}
func (self *CRDTLogootOpBasedDag) Remove(previousLine LogootLine) string {

	newNode := CreateDagNode(Operation{Elem: Element(previousLine), Op: DELETE}, self.GetSys().IpfsNode.Identity.Pretty())
	for dependency := range self.dag.Root_nodes {
		newNode.DagNode.DirectDependency = append(newNode.DagNode.DirectDependency, self.dag.Root_nodes[dependency])
	}

	strFile := self.dag.NextFileName()
	if _, err := os.Stat(strFile); !errors.Is(err, os.ErrNotExist) {
		os.Remove(strFile)
	}
	newNode.ToFile(strFile)
	bytes, err := os.ReadFile(strFile)
	if err != nil {
		panic(fmt.Errorf("ERROR INCREMENT CRDTSetOpBasedDag, could not read file\nerror: %s", err))
	}
	path, err := self.callAddToIPFS(bytes, strFile)
	if err != nil {
		panic(fmt.Errorf("CRDTSetOpBasedDag Decrement, could not add the file to IFPS\nerror: %s", err))
	}

	encodedCid := self.dag.EncodeCid(path)
	c := cid.Cid{}
	err = json.Unmarshal(encodedCid.Str, &c)
	if err != nil {
		panic(fmt.Errorf("CRDTSetOpBasedDag Increment, could not UnMarshal\nerror: %s", err))
	}

	// _, c, _ := cid.CidFromBytes(encodedCid.Str)
	// fmt.Println("encodedCid Decrement :", c.String())
	var pl CRDTDag.CRDTDagNodeInterface = &newNode
	self.dag.AddNode(encodedCid, &pl)
	self.SendRemoteUpdates()
	self.GetDag().UpdateRootNodeFolder()
	self.Data = self.Lookup()
	return c.String()
}

func Create_CRDTLogootOpBasedDag(sys *IpfsLink.IpfsLink, cfg Config.IM_CRDTConfig) CRDTLogootOpBasedDag {
	fmt.Printf("1\n")
	man := CRDTDag.Create_CRDTManager(sys, cfg.PeerName, cfg.BootstrapPeer, cfg.Encode, cfg.Measurement)
	fmt.Printf("2\n")
	crdtSet := CRDTLogootOpBasedDag{dag: man, measurement: cfg.Measurement, Data: Create_CRDTLogootOpBased(sys)}
	fmt.Printf("3\n")
	if cfg.BootstrapPeer == "" {
		x, err := os.ReadFile("initial_value")
		if err != nil {
			panic(fmt.Errorf("Could not read initial_value, error : %s", err))
		}
		fmt.Printf("4\n")

		last := crdtSet.Data.txt.GetFirstNode()
		next := crdtSet.Data.txt.GetLastNode()
		fmt.Printf("5\n") // 6 Never Happens
		generatedPosition := generateLinePositions(last.Line.Pos.pos, next.Line.Pos.pos, 1, sys.IpfsNode.Identity.Pretty())
		fmt.Printf("5.1\n") // 6 Never Happens
		op := Operation{Elem: Element(LogootLine{Pos: PositionIdentifier{clock: time.Now().Nanosecond(), pos: generatedPosition[0]}, Text: string(x)}), Op: INSERT}
		fmt.Printf("5.2\n") // 6 Never Happens
		newNode := CreateDagNode(Operation{Elem: op.Elem, Op: INSERT}, crdtSet.GetSys().IpfsNode.Identity.Pretty())
		fmt.Printf("5.3\n") // 6 Never Happens
		strFile := crdtSet.dag.NextFileName()

		fmt.Printf("6\n")
		if _, err := os.Stat(strFile); !errors.Is(err, os.ErrNotExist) {
			os.Remove(strFile)
		}
		newNode.ToFile(strFile)

		fmt.Printf("7\n")
		bytes, err := os.ReadFile(strFile)
		if err != nil {
			panic(fmt.Errorf("ERROR INCREMENT CRDTSetOpBasedDag, could not read file\nerror: %s", err))
		}
		fmt.Printf("8\n")
		path, err := man.AddToIPFS(crdtSet.dag.Sys, bytes) // Add Inital State ( so it isn't counted as messages)
		if err != nil {
			panic(fmt.Errorf("CRDTSetOpBasedDag Increment, could not add the file to IFPS\nerror: %s", err))
		}

		fmt.Printf("9\n")
		encodedCid := crdtSet.dag.EncodeCid(path)
		c := cid.Cid{}
		err = json.Unmarshal(encodedCid.Str, &c)
		if err != nil {
			panic(fmt.Errorf("CRDTSetOpBasedDag Increment, could not UnMarshal\nerror: %s", err))
		}
		// fmt.Println("encodedCid Increment :", c.String())
		var pl1 CRDTDag.CRDTDagNodeInterface = &newNode

		fmt.Printf("10\n")
		crdtSet.dag.AddNode(encodedCid, &pl1) // TODOSetCrdt Complete Node interface
		crdtSet.Data = crdtSet.Lookup()

	}
	var pl CRDTDag.CRDTDag = &crdtSet

	fmt.Printf("11\n")
	CRDTDag.CheckForRemoteUpdates(&pl, sys.Cr.Sub, man.Sys.Ctx)

	fmt.Printf("12\n")
	return crdtSet
}

func (self *CRDTLogootOpBasedDag) GetSys() *IpfsLink.IpfsLink {

	return self.dag.Sys
}

func (self *CRDTLogootOpBasedDag) Lookup_ToSpecifyType() *CRDT.CRDT {

	crdt := Create_CRDTLogootOpBased(self.GetSys())

	// crdt := CRDTLogootOpBased{
	// 	sys: self.GetSys(),
	// 	txt: LogootText{},
	// }
	for x := range self.dag.GetAllNodes() {
		node := self.dag.GetAllNodesInterface()[x]
		pl := (*(*node).GetEvent()).(*PayloadOpBased)
		if pl.Op.Op == INSERT {
			// fmt.Println("add")
			crdt.Insert(pl.Op.Elem.Pos, pl.Op.Elem.Text)
		} else {
			// fmt.Println("remove")
			crdt.Delete((*(*node).GetEvent()).(*PayloadOpBased).Op.Elem.Pos)
		}
	}
	var pl CRDT.CRDT = &crdt
	return &pl
}
func (self *CRDTLogootOpBasedDag) Lookup() CRDTLogootOpBased {

	// crdt := self.logokup_ToSpecifyType()
	// var pl CRDTDag.CRDTDag = &crdtSet
	return *(*self.Lookup_ToSpecifyType()).(*CRDTLogootOpBased)
}

type TimeTuple struct {
	Cid            string
	RetrievalAlone int
	RetrievalTotal int
	CalculTime     int
	Time_add       int
	Time_encrypt   int
	Time_decrypt   int
	ArrivalTime    int
}

// semaphore usage
func getSema(sema *semaphore.Weighted, ctx context.Context) {
	t := time.Now()
	err := sema.Acquire(ctx, 1)
	for err != nil && time.Since(t) < 10*time.Second {
		time.Sleep(10 * time.Microsecond)
		err = sema.Acquire(ctx, 1)
	}
	if err != nil {
		panic(fmt.Errorf("Semaphore of READ/WRITE file locked !!!!\n Cannot acquire it\n"))
	}
}

func returnSema(sema *semaphore.Weighted) {
	sema.Release(1)
}

// Check update function retrieve files from ipfs (long)
// and then reserves the semaphore to actually modify the data (short)
func (self *CRDTLogootOpBasedDag) CheckUpdate(sema *semaphore.Weighted) []TimeTuple {
	received := make([]TimeTuple, 0)
	files, err := ioutil.ReadDir(self.GetDag().Nodes_storage_enplacement + "/remote")
	if err != nil {
		fmt.Printf("CheckUpdate - Checkupdate could not open folder\nerror: %s\n", err)
	} else {
		ti := time.Now()
		to_add := make([]([]byte), 0)
		computetime := make([]int64, 0)
		arrivalTime := make([]int64, 0)
		for _, file := range files {
			if file.Size() > 0 && !strings.Contains(file.Name(), ".ArrivalTime") {
				fil, err := os.OpenFile(self.GetDag().Nodes_storage_enplacement+"/remote/"+file.Name(), os.O_RDONLY, os.ModeAppend)
				if err != nil {
					panic(fmt.Errorf("error in checkupdate, Could not open the sub file\nError: %s", err))
				}
				stat, err := fil.Stat()
				if err != nil {
					panic(fmt.Errorf("error in checkupdate, Could not get stat the sub file\nError: %s", err))
				}
				bytesread := make([]byte, stat.Size())
				n, err := fil.Read(bytesread)
				if err != nil {
					panic(fmt.Errorf("error in checkupdate, Could not read the sub file\nError: %s", err))
				}

				// fmt.Println("stat.size :", stat.Size(), "read :", n)
				if int64(n) != stat.Size() {
					panic(fmt.Errorf("error in checkupdate, Could not read entirely the sub file\nError: read %d byte unstead of %d", n, stat.Size()))
				}
				err = fil.Close()
				if err != nil {
					panic(fmt.Errorf("error in checkupdate, Could not close the sub file\nError: %s", err))
				}
				if !self.IsKnown(CRDTDag.EncodedStr{Str: bytesread}) {
					to_add = append(to_add, bytesread)
				}
				s := cid.Cid{}
				json.Unmarshal(bytesread, &s)

				err = os.Remove(self.GetDag().Nodes_storage_enplacement + "/remote/" + file.Name())
				if err != nil || errors.Is(err, os.ErrNotExist) {
					panic(fmt.Errorf("error in checkupdate, Could not remove the sub file\nError: %s", err))
				}

				// Take the time measurement of this file
				// Get the time of arrival to compute pubsub time
				fil, err = os.OpenFile(self.GetDag().Nodes_storage_enplacement+"/remote/"+file.Name()+".ArrivalTime", os.O_RDONLY, os.ModeAppend)
				if err != nil {
					panic(fmt.Errorf("error in checkupdate, Could not open the sub file\nError: %s", err))
				}
				stat, err = fil.Stat()
				if err != nil {
					panic(fmt.Errorf("error in checkupdate, Could not get stat the sub file\nError: %s", err))
				}
				bytesread = make([]byte, stat.Size())
				n, err = fil.Read(bytesread)
				if err != nil {
					panic(fmt.Errorf("error in checkupdate, Could not read the sub file\nError: %s", err))
				}

				// fmt.Println("stat.size :", stat.Size(), "read :", n)
				if int64(n) != stat.Size() {
					panic(fmt.Errorf("error in checkupdate, Could not read entirely the sub file\nError: read %d byte unstead of %d", n, stat.Size()))
				}
				err = fil.Close()
				if err != nil {
					panic(fmt.Errorf("error in checkupdate, Could not close the sub file\nError: %s", err))
				}
				time_of_arrival, _ := strconv.Atoi(string(bytesread))
				arrivalTime = append(arrivalTime, int64(time_of_arrival))

				//computation time, time to manage this file
				timeToCompute := time.Since(ti).Nanoseconds()
				computetime = append(computetime, timeToCompute)
				ti = time.Now()
			} else {
				fmt.Printf("Remote folder contain a FILE of a NULL SIZE\n")
			}
		}

		// apply the update on the peer's data
		getSema(sema, self.GetSys().Ctx)
		received = self.add_cids(to_add, computetime, arrivalTime, ti)

		if len(to_add) > 0 {
			self.GetDag().UpdateRootNodeFolder()
		}
		self.Data = self.Lookup()
		returnSema(sema)
	}
	return received
}

func (self *CRDTLogootOpBasedDag) add_cids(to_add []([]byte), computetime []int64, arrivalTime []int64, ti time.Time) []TimeTuple {
	received := make([]TimeTuple, 0)

	bytes_encoded := make([]CRDTDag.EncodedStr, 0)

	for _, bytesread := range to_add {
		bytes_encoded = append(bytes_encoded, CRDTDag.EncodedStr{Str: bytesread})
	}

	filesWritten := self.Merge(bytes_encoded)

	for index, bytesread := range to_add {
		s := cid.Cid{}
		json.Unmarshal(bytesread, &s)
		timeRetrieve := 0
		timeDecrypt := 0
		if self.measurement && filesWritten[index] != "node1/node1" {
			// Get Time of Retrieval
			str, err := os.ReadFile(filesWritten[index] + ".timeRetrieve")
			if err != nil {
				panic(fmt.Errorf("Set.go - Could not read Time to retrieve measurement\nError: %s", err))
			}
			timeRetrieve, err = strconv.Atoi(string(str))
			if err != nil {
				panic(fmt.Errorf("Set.go - Could not translate Time to retrieve to string, maybe malformerd ?\nError: %s", err))
			}

			err = os.Remove(filesWritten[index] + ".timeRetrieve")
			if err != nil {
				panic(fmt.Errorf("Set.go - Could not remove Time to retrieve file\nError: %s", err))
			}

			//If we use it, get time of decryption of the file
			if self.dag.Key != "" {
				str, err = os.ReadFile(filesWritten[index] + ".timeDecrypt")
				if err != nil {
					panic(fmt.Errorf("Set.go - Could not read Time to decrypt measurement\nError: %s", err))
				}
				timeDecrypt, err = strconv.Atoi(string(str))
				if err != nil {
					panic(fmt.Errorf("Set.go - Could not translate Time to retrieve to string, maybe malformerd ?\nError: %s", err))
				}
				err = os.Remove(filesWritten[index] + ".timeDecrypt")
				if err != nil {
					panic(fmt.Errorf("Set.go - Could not remove Time to Decrypt file\nError: %s", err))
				}

			}

		}
		// fmt.Println("calling UpdateRootNodeFolder")

		received = append(received, TimeTuple{Cid: s.String(), RetrievalAlone: timeRetrieve, RetrievalTotal: timeRetrieve * len(to_add), CalculTime: int(computetime[index]), ArrivalTime: int(arrivalTime[index]), Time_decrypt: timeDecrypt, Time_encrypt: 0})
	}

	self.GetDag().UpdateRootNodeFolder()
	return received
}

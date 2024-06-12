package Set

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

type Element string
type OpNature int

const (
	ADD OpNature = iota
	REMOVE
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

type CRDTSetOpBased struct {
	sys     *IpfsLink.IpfsLink
	added   []string
	removed []string
}

func Create_CRDTSetOpBased(s *IpfsLink.IpfsLink) CRDTSetOpBased {
	return CRDTSetOpBased{
		sys:     s,
		added:   make([]string, 0),
		removed: make([]string, 0),
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
func (self *CRDTSetOpBased) Add(x string) {
	if search(self.added, x) == -1 {
		self.added = append(self.added, x)
	}
}

func (self *CRDTSetOpBased) Remove(x string) {
	if search(self.removed, x) == -1 {
		self.removed = append(self.removed, x)
	}
}

func (self *CRDTSetOpBased) Lookup() []string {
	i := make([]string, 0)
	fmt.Println("size", len(self.added))
	for x := range self.added {
		if search(self.removed, self.added[x]) == -1 {
			i = append(i, self.added[x])
			i = append(i, ",")
		}
	}

	return i
}

func (self *CRDTSetOpBased) ToFile(file string) {

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

type CRDTSetOpBasedDagNode struct {
	DagNode CRDTDag.CRDTDagNode
}

func (self *CRDTSetOpBasedDagNode) FromFile(fil string) {
	var pl Payload.Payload = &PayloadOpBased{}
	self.DagNode.CreateNodeFromFile(fil, &pl)
}

func (self *CRDTSetOpBasedDagNode) GetDirect_dependency() []CRDTDag.EncodedStr {

	return self.DagNode.DirectDependency
}

func (self *CRDTSetOpBasedDagNode) ToFile(file string) {

	self.DagNode.ToFile(file)
}
func (self *CRDTSetOpBasedDagNode) GetEvent() *Payload.Payload {

	return self.DagNode.Event
}
func (self *CRDTSetOpBasedDagNode) GetPiD() string {

	return self.DagNode.PID
}
func (self *CRDTSetOpBasedDagNode) CreateEmptyNode() *CRDTDag.CRDTDagNodeInterface {
	n := CreateDagNode(Operation{}, "")
	var node CRDTDag.CRDTDagNodeInterface = &n
	return &node
}
func CreateDagNode(o Operation, id string) CRDTSetOpBasedDagNode {
	var pl Payload.Payload = &PayloadOpBased{Op: o, Id: id}
	slic := make([]CRDTDag.EncodedStr, 0)
	return CRDTSetOpBasedDagNode{
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

type CRDTSetOpBasedDag struct {
	dag         CRDTDag.CRDTManager
	measurement bool
}

func (self *CRDTSetOpBasedDag) GetDag() *CRDTDag.CRDTManager {

	return &self.dag
}
func (self *CRDTSetOpBasedDag) SendRemoteUpdates() {

	self.dag.SendRemoteUpdates()
}
func (self *CRDTSetOpBasedDag) GetCRDTManager() *CRDTDag.CRDTManager {

	return &self.dag
}
func (self *CRDTSetOpBasedDag) IsKnown(cid CRDTDag.EncodedStr) bool {

	find := false
	for x := range self.dag.GetAllNodes() {
		if string(self.dag.GetAllNodes()[x]) == string(cid.Str) {
			find = true
			break
		}
	}
	return find
}
func (self *CRDTSetOpBasedDag) Merge(cids []CRDTDag.EncodedStr) []string {

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
	return fils
}

func (self *CRDTSetOpBasedDag) remoteAddNode(cID CRDTDag.EncodedStr, newnode CRDTSetOpBasedDagNode) {
	var pl CRDTDag.CRDTDagNodeInterface = &newnode
	self.dag.RemoteAddNodeSuper(cID, &pl)
}

func (self *CRDTSetOpBasedDag) callAddToIPFS(bytes []byte, file string) (path.Resolved, error) {
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

func (self *CRDTSetOpBasedDag) Add(x string) (string, TimeTuple) {
	newNode := CreateDagNode(Operation{Elem: Element(x), Op: ADD}, self.GetSys().IpfsNode.Identity.Pretty())
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

	return c.String(), times
}
func (self *CRDTSetOpBasedDag) Remove(x string) string {

	newNode := CreateDagNode(Operation{Elem: Element(x), Op: REMOVE}, self.GetSys().IpfsNode.Identity.Pretty())
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
	return c.String()
}

func Create_CRDTSetOpBasedDag(sys *IpfsLink.IpfsLink, cfg Config.IM_CRDTConfig) CRDTSetOpBasedDag {

	man := CRDTDag.Create_CRDTManager(sys, cfg.PeerName, cfg.BootstrapPeer, cfg.Encode, cfg.Measurement)
	crdtSet := CRDTSetOpBasedDag{dag: man, measurement: cfg.Measurement}
	if cfg.BootstrapPeer == "" {
		x, err := os.ReadFile("initial_value")
		if err != nil {
			panic(fmt.Errorf("Could not read initial_value, error : %s", err))
		}
		newNode := CreateDagNode(Operation{Elem: Element(x), Op: ADD}, crdtSet.GetSys().IpfsNode.Identity.Pretty())
		strFile := crdtSet.dag.NextFileName()

		if _, err := os.Stat(strFile); !errors.Is(err, os.ErrNotExist) {
			os.Remove(strFile)
		}
		newNode.ToFile(strFile)

		bytes, err := os.ReadFile(strFile)
		if err != nil {
			panic(fmt.Errorf("ERROR INCREMENT CRDTSetOpBasedDag, could not read file\nerror: %s", err))
		}
		path, err := man.AddToIPFS(crdtSet.dag.Sys, bytes) // Add Inital State ( so it isn't counted as messages)
		if err != nil {
			panic(fmt.Errorf("CRDTSetOpBasedDag Increment, could not add the file to IFPS\nerror: %s", err))
		}

		encodedCid := crdtSet.dag.EncodeCid(path)
		c := cid.Cid{}
		err = json.Unmarshal(encodedCid.Str, &c)
		if err != nil {
			panic(fmt.Errorf("CRDTSetOpBasedDag Increment, could not UnMarshal\nerror: %s", err))
		}
		// fmt.Println("encodedCid Increment :", c.String())
		var pl1 CRDTDag.CRDTDagNodeInterface = &newNode

		crdtSet.dag.AddNode(encodedCid, &pl1) // TODOSetCrdt Complete Node interface

	}
	var pl CRDTDag.CRDTDag = &crdtSet

	CRDTDag.CheckForRemoteUpdates(&pl, sys.Cr.Sub, man.Sys.Ctx)

	return crdtSet
}

func (self *CRDTSetOpBasedDag) GetSys() *IpfsLink.IpfsLink {

	return self.dag.Sys
}

func (self *CRDTSetOpBasedDag) Lookup_ToSpecifyType() *CRDT.CRDT {

	crdt := CRDTSetOpBased{
		sys:     self.GetSys(),
		added:   make([]string, 0),
		removed: make([]string, 0),
	}
	for x := range self.dag.GetAllNodes() {
		node := self.dag.GetAllNodesInterface()[x]
		if (*(*node).GetEvent()).(*PayloadOpBased).Op.Op == ADD {
			// fmt.Println("add")
			crdt.Add(string((*(*node).GetEvent()).(*PayloadOpBased).Op.Elem))
		} else {
			// fmt.Println("remove")
			crdt.Remove(string((*(*node).GetEvent()).(*PayloadOpBased).Op.Elem))
		}
	}
	var pl CRDT.CRDT = &crdt
	return &pl
}
func (self *CRDTSetOpBasedDag) Lookup() CRDTSetOpBased {

	// crdt := self.logokup_ToSpecifyType()
	// var pl CRDTDag.CRDTDag = &crdtSet
	return *(*self.Lookup_ToSpecifyType()).(*CRDTSetOpBased)
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
func (self *CRDTSetOpBasedDag) CheckUpdate(sema *semaphore.Weighted) []TimeTuple {
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

		returnSema(sema)
	}
	return received
}

func (self *CRDTSetOpBasedDag) add_cids(to_add []([]byte), computetime []int64, arrivalTime []int64, ti time.Time) []TimeTuple {
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

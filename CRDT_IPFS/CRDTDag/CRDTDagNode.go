package CRDTDag

import (
	"IPFS_CRDT/Payload"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type CRDTDagNodeInterface interface {
	ToFile(file string)
	FromFile(file string)

	GetEvent() *Payload.Payload
	GetPiD() string

	GetDirect_dependency() []EncodedStr
	CreateEmptyNode() *CRDTDagNodeInterface
}

type CRDTDagNode struct {
	Event            *Payload.Payload
	PID              string
	DirectDependency []EncodedStr
}

func (this *CRDTDagNode) CreateNode(dd []EncodedStr, peerID string, op *Payload.Payload) {

	this.Event = op
	this.PID = peerID
	for x := range dd {
		this.DirectDependency = append(this.DirectDependency, dd[x])
	}
}

func (this *CRDTDagNode) ToString() string {

	s := "peerID:" + this.PID + "\n"
	s += "op:" + (*this.Event).ToString() + "\n"
	for x := range this.DirectDependency {
		msgBytes, err := json.Marshal(this.DirectDependency[x].Str)
		if err != nil {
			panic(fmt.Errorf("encoding dependency issue\nError: %s", err))
		}
		s += string(msgBytes) + " "
	}
	s += "\n"
	return s
}

func (this *CRDTDagNode) ToFile(file string) {

	f, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		panic(fmt.Errorf("CRDTDagNode - ToFile Could not open the file %s\nError: %s", file, err))
	}
	defer f.Close()

	a, err := json.Marshal(this)
	if err != nil {
		panic(fmt.Errorf("CRDTDagNode - ToFile Could not translate CRDTDagNode to json\nError: %s", err))
	}

	defer f.Close()
	if _, err = f.Write(a); err != nil {
		panic(fmt.Errorf("CRDTDagNode - ToFile Could not write to the file %s\nError: %s", file, err))
	}
}

func (this *CRDTDagNode) CreateNodeFromFile(file string, p *Payload.Payload) {

	data, err := ioutil.ReadFile(file)
	if err != nil {
		panic(fmt.Errorf("CRDTDagNode - CreateFromFile Could not Read the file %s\nError: %s", file, err))
	}
	err = json.Unmarshal(data, this)
	if err != nil {
		panic(fmt.Errorf("CRDTDagNode - UnmmarshalDidn't work, FileName: %s\nError: %s", file, err))
	}
	// this.Event = p
	// s := string(data)
	// str := ""
	// step := 0
	// for k := range s {
	// 	if step == 0 {
	// 		if s[k] == '\n' {
	// 			this.PID = str
	// 			step += 1
	// 			str = ""
	// 		} else if str == "peerID:" {
	// 			str = "" + string(s[k])
	// 		} else {
	// 			str += string(s[k])
	// 		}
	// 	} else if step == 1 {
	// 		if s[k] == '\n' {
	// 			(*this.Event).FromString(str)
	// 			// fmt.Println("Event From string : ", str, "from the message", s, "\nthe event :", (*this.Event).ToString())
	// 			step += 1
	// 			str = ""
	// 		} else if str == "op:" {
	// 			str = "" + string(s[k])
	// 		} else {
	// 			str += string(s[k])
	// 		}
	// 	} else if step == 2 {
	// 		if s[k] == '\n' {
	// 			step += 1
	// 		} else if s[k] == ' ' {
	// 			msgBytes := ""
	// 			err := json.Unmarshal([]byte(str), &msgBytes)
	// 			if err != nil {
	// 				panic(fmt.Errorf("CreateNodeFromFile Could not Unmarshall message\nError: %s", err))
	// 			}
	// 			this.DirectDependency = append(this.DirectDependency, EncodedStr{[]byte(msgBytes)})
	// 			str = ""
	// 		} else {
	// 			str += string(s[k])
	// 		}
	// 	}

	// }
}

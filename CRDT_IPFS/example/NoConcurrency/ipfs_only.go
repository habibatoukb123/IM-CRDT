package NoConcurrency

import (
	IpfsLink "IPFS_CRDT/ipfsLink"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/ipfs/go-cid"
	Files "github.com/ipfs/go-libipfs/files"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

// =======================================================================================
// StrSet
// =======================================================================================

type StrSet struct {
	Set                 []string
	sys                 *IpfsLink.IpfsLink
	storage_emplacement string
	nextNodeName        int
	last_name           string
	bootstrap           bool
}

func (self *StrSet) NextFileName() string {
	if !self.bootstrap {
		os.Remove(self.last_name)
	}
	res := self.storage_emplacement + "/node" + strconv.Itoa(self.nextNodeName)
	self.nextNodeName += 1
	self.last_name = res
	return res
}

func (self *StrSet) NextRemoteFileName() string {

	res := self.storage_emplacement + "/remote/" + strconv.Itoa(self.nextNodeName)
	self.nextNodeName += 1
	return res
}
func (self *StrSet) CheckForRemoteUpdates(sub *pubsub.Subscription, c context.Context) {
	go func() {
		for {
			msg, err := sub.Next(c)
			if err != nil {

				panic(fmt.Errorf("Check For remote update failed, message not received\nError: %s", err))
			} else if msg.ReceivedFrom == (*self).sys.Cr.Host.ID() {
				fmt.Println("Received message from myself")
				continue
			} else {
				fmt.Println("Received message from", msg.ReceivedFrom,
					"data:", string(msg.Data))
				fileName := (*self).NextRemoteFileName()
				if _, err := os.Stat(fileName); !errors.Is(err, os.ErrNotExist) {
					os.Remove(fileName)
				}
				fil, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY, 0755)
				if err != nil {
					panic(fmt.Errorf("error in checkForRemoteUpdate, Could not open the sub file\nError: %s", err))
				}
				_, err = fil.Write(msg.GetData())
				if err != nil {
					panic(fmt.Errorf("error in checkForRemoteUpdate, Could not write the sub file\nError: %s", err))
				}
				err = fil.Close()
				if err != nil {
					panic(fmt.Errorf("error in checkForRemoteUpdate, Could not close the sub file\nError: %s", err))
				}
			}
		}
	}()
}
func InitIPFSSet(sys *IpfsLink.IpfsLink, storage_emplacement string, bootStrapPeer string) *StrSet {
	man := createIpfsSet(sys, storage_emplacement, bootStrapPeer)

	man.CheckForRemoteUpdates(sys.Cr.Sub, man.sys.Ctx)
	return &man
}

func createIpfsSet(sys *IpfsLink.IpfsLink, sr string, bootStrapPeer string) StrSet {
	v := StrSet{
		Set:                 make([]string, 0),
		sys:                 sys,
		storage_emplacement: sr,
		nextNodeName:        0,
	}

	v.bootstrap = (bootStrapPeer == "")

	x, err := os.ReadFile("initial_value")
	if err != nil {
		panic(fmt.Errorf("Could not read initial_value, error : %s", err))
	}
	v.Set = append(v.Set, string(x))
	return v

}
func (self *StrSet) SendFile() []byte {
	b, err := json.Marshal(self.Set)
	if err != nil {
		panic(fmt.Errorf("could not marshal in sendFile StrSet\nError: %s", err))
	}
	path, err := IpfsLink.AddIPFS(self.sys, b)
	if err != nil {
		panic(fmt.Errorf("could not add in IPFS in sendFile StrSet\nError: %s", err))
	}
	by, err := json.Marshal(path.Cid())
	if err != nil {
		panic(fmt.Errorf("could notmarshal in IPFS in sendFile StrSet\nError: %s", err))
	}
	IpfsLink.PubIPFS(self.sys, by)
	strFile := self.NextFileName()
	if _, err := os.Stat(strFile); !errors.Is(err, os.ErrNotExist) {
		os.Remove(strFile)
	}
	file, err := os.OpenFile(strFile, os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		panic(fmt.Errorf("could not open file in sendFile StrSet\nError: %s", err))
	}
	file.Write(b)
	err = file.Close()
	if err != nil {
		panic(fmt.Errorf("could not close file in sendFile StrSet\nError: %s", err))
	}
	return []byte(path.Cid().String())
}
func search(list []string, x string) int {
	for i := 0; i < len(list); i++ {
		if list[i] == x {
			return i
		}
	}
	return -1
}
func (self *StrSet) Add(x string) []byte {
	if search(self.Set, x) == -1 {
		self.Set = append(self.Set, x)
	}
	return self.SendFile()
}
func (self *StrSet) Remove(x string) []byte {
	if i := search(self.Set, x); i != -1 {
		self.Set[i] = self.Set[len(self.Set)-1]
		self.Set = self.Set[:len(self.Set)-1]
	}
	return self.SendFile()
}

func (self *StrSet) CheckUpdate() []string {
	files, err := ioutil.ReadDir(self.storage_emplacement + "/remote")

	received := make([]string, 0)
	if err != nil {
		fmt.Printf("CheckUpdate - Checkupdate could not open folder\nerror: %s\n", err)
	} else {
		for _, file := range files {
			if file.Size() > 0 {
				fil, err := os.OpenFile(self.storage_emplacement+"/remote/"+file.Name(), os.O_RDONLY, os.ModeAppend)
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
				ccid := cid.Cid{}
				err = json.Unmarshal(bytesread, &(ccid))
				if err != nil {
					fmt.Println(string(bytesread))
					panic(fmt.Errorf("error in checkupdate, Could not Unmarshall\nError: %s", err))
				}
				received = append(received, ccid.String())

				newFiles, err := IpfsLink.GetIPFS(self.sys, append(make([]cid.Cid, 0), ccid))
				if err != nil {
					panic(fmt.Errorf("issue retrieving the IPFS Node :%s", err))
				}
				newNodeFile := self.NextFileName()
				Files.WriteTo(newFiles[0], newNodeFile)
				// os.WriteFile(newNodeFile, newFiles[0].Node.RawData(), 0644)

				fileNotFinished := true
				for fileNotFinished {
					fileNotFinished = false
					fil, err = os.OpenFile(newNodeFile, os.O_RDONLY, os.ModeAppend)
					if err != nil {
						panic(fmt.Errorf("2error in checkupdate, Could not open the sub file\nError: %s", err))
					}
					stat, err = fil.Stat()
					if err != nil {
						panic(fmt.Errorf("2error in checkupdate, Could not get stat the sub file\nError: %s", err))
					}
					bytesread = make([]byte, stat.Size())
					n, err = fil.Read(bytesread)
					if err != nil {
						panic(fmt.Errorf("2error in checkupdate, Could not read the sub file\nError: %s", err))
					}

					err = json.Unmarshal(bytesread, &(self.Set))
					if err != nil || errors.Is(err, os.ErrNotExist) {
						fileNotFinished = true
						fmt.Fprintf(os.Stderr, "2error in checkupdate, Could not unmarshall \nError: %s", err)
					}
				}

				err = os.Remove(self.storage_emplacement + "/remote/" + file.Name())
				if err != nil || errors.Is(err, os.ErrNotExist) {
					panic(fmt.Errorf("error in checkupdate, Could not remove the sub file\nError: %s", err))
				}

			} else {
				fmt.Printf("FILE SIZE NULL")
			}
		}
	}
	return received
}

func Create_StrSet(s *IpfsLink.IpfsLink, str string) StrSet {
	return StrSet{Set: make([]string, 0), sys: s, storage_emplacement: str, nextNodeName: 0}
}

func (self *StrSet) Lookup() []string {
	return self.Set
}

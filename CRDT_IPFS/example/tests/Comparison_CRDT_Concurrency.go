package tests

import (
	"IPFS_CRDT/Config"
	Set "IPFS_CRDT/example/2PSet"
	IpfsLink "IPFS_CRDT/ipfsLink"
	"io/ioutil"

	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"golang.org/x/sync/semaphore"
	// "github.com/beevik/ntp"
)

func GetTime(ntpServ string) int {
	return int(time.Now().UnixNano())
}

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

// \/ BOOTSTRAP PEER IS THIS ONE \/
func Peer1Concu(cfg Config.IM_CRDTConfig) {
	fileRead, err := os.OpenFile(cfg.PeerName+"/time/FileRead.log", os.O_CREATE|os.O_WRONLY, 0755)
	file, err := os.OpenFile(cfg.PeerName+"/time/time.csv", os.O_CREATE|os.O_WRONLY, 0755)
	sema := semaphore.NewWeighted(1)

	sys1, err := IpfsLink.InitNode(cfg.PeerName, "", make([]byte, 0), cfg.SwarmKey, cfg.ParallelRetrieve)
	if err != nil {
		panic(fmt.Errorf("failed To instanciate IFPS & LibP2P clients : %s", err))
	}

	str := ""
	for i := range sys1.Cr.Host.Addrs() {
		s := sys1.Cr.Host.Addrs()[i].String()
		str += s + "/p2p/" + sys1.Cr.Host.ID().String() + "\n"
	}
	if _, err := os.Stat("./ID2"); !errors.Is(err, os.ErrNotExist) {
		os.Remove("./ID2")
	}

	WriteFile("./ID2", []byte(str))

	IpfsLink.WritePeerInfo(*sys1, "./IDBootstrapIPFS")

	time.Sleep(20 * time.Second)

	getSema(sema, sys1.Ctx)
	SetCrdt1 := Set.Create_CRDTSetOpBasedDag(sys1, cfg)
	returnSema(sema)

	fileRead.WriteString("Taking Sema to write headers ... ")
	getSema(sema, sys1.Ctx)
	file.WriteString("CID,time,time_retrieve,time_compute,time_add_IPFS,time_encrypt,time_decrypt,time_Retreive_Whole_Batch,ArrivalTime,TimeLookupFolderBatch,TimeRemoveFilesBatch\n")
	returnSema(sema)
	fileRead.WriteString("Header just written\n")
	if err != nil {
		panic(fmt.Errorf("Error openning file file\nerror : %s", err))
	}
	fmt.Println("Starting the Set, sleeping 30s to wait others")

	ti := time.Now()
	var strList []Set.TimeTuple
	// Sleep 60s before emiting updates to wait others
	for time.Since(ti) < 60*time.Second {
		strList = make([]Set.TimeTuple, 0)
		if cfg.TestMode {
			files, err := ioutil.ReadDir(SetCrdt1.GetDag().Nodes_storage_enplacement + "/remote")
			if err != nil {
				fmt.Printf("CheckUpdate - Checkupdate could not open folder\nerror: %s\n", err)
			}
			if len(files) >= 2*20 {
				fileRead.WriteString("CheckUpdates 20 Verison ! assure\n= = = = = = =\n")
				time.Sleep(time.Duration(cfg.WaitTime) * time.Microsecond)
				strList = SetCrdt1.CheckUpdate_20Version(sema)
			}
		} else {
			time.Sleep(time.Duration(cfg.WaitTime) * time.Microsecond)
			strList = SetCrdt1.CheckUpdate(sema)
		}
		if len(strList) > 0 {
			fileRead.WriteString("Just Received some updates\n")
			t := strconv.Itoa(GetTime(cfg.NtpServ))

			for j := 0; j < len(strList); j++ {
				getSema(sema, sys1.Ctx)
				file.WriteString(strList[j].Cid + "," + t + "," + strconv.Itoa(strList[j].RetrievalAlone) + "," + strconv.Itoa(strList[j].CalculTime) + ",0,0," + strconv.Itoa(strList[j].Time_decrypt) + "," + strconv.Itoa(strList[j].RetrievalTotal) + "," + strconv.Itoa(strList[j].ArrivalTime) + "," + strconv.Itoa(strList[j].TimeLookupFolder) + "," + strconv.Itoa(strList[j].TimeRemoveFiles) + "\n")
				returnSema(sema)
				fileRead.WriteString("writing 1 line\n")
			}
			fileRead.WriteString("all update received are handled\n= = = = = = =\n")
		}
	}

	fmt.Printf("Starting the Set, updating %d times\n", cfg.UpdatesNB)
	ti = time.Now()

	// Send updates concurrently every 1 seconds
	go sendUpdates(cfg.UpdatesNB, &SetCrdt1, cfg.NtpServ, file, sys1.Cr.Id, sema, cfg)

	//regularly scan files if there is any new received updates
	for {
		strList = make([]Set.TimeTuple, 0)
		if cfg.TestMode {
			files, err := ioutil.ReadDir(SetCrdt1.GetDag().Nodes_storage_enplacement + "/remote")
			if err != nil {
				fmt.Printf("CheckUpdate - Checkupdate could not open folder\nerror: %s\n", err)
			}
			if len(files) >= 2*20 {
				fileRead.WriteString("CheckUpdates 20 Verison ! assure\n= = = = = = =\n")
				time.Sleep(time.Duration(cfg.WaitTime) * time.Microsecond)
				strList = SetCrdt1.CheckUpdate_20Version(sema)
			}
		} else {
			time.Sleep(time.Duration(cfg.WaitTime) * time.Microsecond)
			strList = SetCrdt1.CheckUpdate(sema)
		}
		if len(strList) > 0 {
			fileRead.WriteString("Just Received some updates\n")
			t := strconv.Itoa(GetTime(cfg.NtpServ))

			for j := 0; j < len(strList); j++ {
				getSema(sema, sys1.Ctx)
				file.WriteString(strList[j].Cid + "," + t + "," + strconv.Itoa(strList[j].RetrievalAlone) + "," + strconv.Itoa(strList[j].CalculTime) + ",0,0," + strconv.Itoa(strList[j].Time_decrypt) + "," + strconv.Itoa(strList[j].RetrievalTotal) + "," + strconv.Itoa(strList[j].ArrivalTime) + "," + strconv.Itoa(strList[j].TimeLookupFolder) + "," + strconv.Itoa(strList[j].TimeRemoveFiles) + "\n")
				returnSema(sema)
				fileRead.WriteString("writing 1 line\n")
			}
			fileRead.WriteString("all update received are handled\n= = = = = = =\n")
		}
		// x := SetCrdt1.Lookup()
		// fmt.Println("New Value of the Set:", x.Lookup())
	}
}

func Peer2Concu(cfg Config.IM_CRDTConfig) {
	IPFSbootstrapBytes, err := os.ReadFile(cfg.IPFSbootstrap)
	sema := semaphore.NewWeighted(1)
	if err != nil {
		panic(fmt.Errorf("failed to read ipfs bootstrap peer multiaddr : %s", err))
	}
	sys1, err := IpfsLink.InitNode(cfg.PeerName, cfg.BootstrapPeer, IPFSbootstrapBytes, cfg.SwarmKey, cfg.ParallelRetrieve)
	if err != nil {
		panic(fmt.Errorf("failed to instanciate ipfs & libp2p clients : %s", err))
	}
	time.Sleep(10 * time.Second)

	getSema(sema, sys1.Ctx)
	SetCrdt1 := Set.Create_CRDTSetOpBasedDag(sys1, cfg)
	returnSema(sema)

	file, err := os.OpenFile(cfg.PeerName+"/time/time.csv", os.O_CREATE|os.O_WRONLY, 0755)
	file.WriteString("CID,time,time_retrieve,time_compute,time_add_IPFS,time_encrypt,time_decrypt,time_Retreive_Whole_Batch,ArrivalTime,TimeLookupFolderBatch,TimeRemoveFilesBatch\n")
	if err != nil {
		panic(fmt.Errorf("error openning file file\nerror : %s", err))
	}
	fileRead, err := os.OpenFile(cfg.PeerName+"/time/FileRead.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0755)
	if err != nil {
		panic(fmt.Errorf("error openning file file\nerror : %s", err))
	}
	fmt.Printf("Starting the Set, updating %d times\n", cfg.UpdatesNB)
	var strList []Set.TimeTuple
	for {
		strList = make([]Set.TimeTuple, 0)
		if cfg.TestMode {
			files, err := ioutil.ReadDir(SetCrdt1.GetDag().Nodes_storage_enplacement + "/remote")
			if err != nil {
				fmt.Printf("CheckUpdate - Checkupdate could not open folder\nerror: %s\n", err)
			}
			if len(files) >= 2*20 {
				time.Sleep(time.Duration(cfg.WaitTime) * time.Microsecond)
				strList = SetCrdt1.CheckUpdate_20Version(sema)
				if len(strList) > 0 {
					fileRead.WriteString(fmt.Sprintf("Timelookup : %d, timeRemove : %d\n", strList[0].TimeLookupFolder, strList[0].TimeRemoveFiles))
					fileRead.WriteString(fmt.Sprintf("strList : %s\n\n", strList[0].Cid))
				}
			}
		} else {
			time.Sleep(time.Duration(cfg.WaitTime) * time.Microsecond)
			strList = SetCrdt1.CheckUpdate(sema)
		}
		if len(strList) > 0 {
			t := strconv.Itoa(GetTime(cfg.NtpServ))

			for j := 0; j < len(strList); j++ {
				file.WriteString(strList[j].Cid + "," + t + "," + strconv.Itoa(strList[j].RetrievalAlone) + "," + strconv.Itoa(strList[j].CalculTime) + ",0,0," + strconv.Itoa(strList[j].Time_decrypt) + "," + strconv.Itoa(strList[j].RetrievalTotal) + "," + strconv.Itoa(strList[j].ArrivalTime) + "," + strconv.Itoa(strList[j].TimeLookupFolder) + "," + strconv.Itoa(strList[j].TimeRemoveFiles) + "\n")
			}
		}

	}
	fileRead.Close()
}

func Peer2ConcuUpdate(cfg Config.IM_CRDTConfig) {
	sema := semaphore.NewWeighted(1)

	// Reading the IPFSBootstrap file
	fileInfo, err := os.Stat(cfg.IPFSbootstrap)
	if err != nil {
		panic(fmt.Errorf("Peer2ConcuUpdate - could Not Open IPFSBootstrap file toread bootstrap address\nerror: %s", err))
	}
	IPFSbootstrapBytes := make([]byte, fileInfo.Size())
	// Writing bytes in the file @file
	fil, err := os.OpenFile(cfg.IPFSbootstrap, os.O_RDONLY, 0755)
	if err != nil {
		panic(fmt.Errorf("Peer2ConcuUpdate - could Not Open IPFSBootstrap file to read it\nerror: %s", err))
	}
	_, err = fil.Read(IPFSbootstrapBytes)
	if err != nil {
		panic(fmt.Errorf("could Not read IPFSBootstrap file - Peer2ConcuUpdate - \nerror: %s", err))
	}
	err = fil.Close()
	if err != nil {
		panic(fmt.Errorf("could Not Close IPFSBootstrap file - Peer2ConcuUpdate\nerror: %s", err))
	}
	sys1, err := IpfsLink.InitNode(cfg.PeerName, cfg.BootstrapPeer, IPFSbootstrapBytes, cfg.SwarmKey, cfg.ParallelRetrieve)
	if err != nil {
		fmt.Printf("Failed To instanciate IFPS & LibP2P clients : %s", err)
		panic(err)
	}
	time.Sleep(10 * time.Second)

	getSema(sema, sys1.Ctx)
	SetCrdt1 := Set.Create_CRDTSetOpBasedDag(sys1, cfg)
	returnSema(sema)
	file, _ := os.OpenFile(cfg.PeerName+"/time/time.csv", os.O_CREATE|os.O_WRONLY, 0755)
	fileRead, err := os.OpenFile(cfg.PeerName+"/time/FileRead.log", os.O_CREATE|os.O_WRONLY, 0755)
	fileRead.WriteString("Taking Sema to write headers ... ")
	getSema(sema, sys1.Ctx)
	file.WriteString("CID,time,time_retrieve,time_compute,time_add_IPFS,time_encrypt,time_decrypt,time_Retreive_Whole_Batch,ArrivalTime,TimeLookupFolderBatch,TimeRemoveFilesBatch\n")
	returnSema(sema)
	fileRead.WriteString("Header just written\n")
	if err != nil {
		fmt.Printf("Error openning file file\nerror : %s", err)
		panic(err)
	}

	// Sleep 60s before emiting updates to wait others
	ti := time.Now()
	var strList []Set.TimeTuple
	for time.Since(ti) < 60*time.Second {
		strList = make([]Set.TimeTuple, 0)
		if cfg.TestMode {
			files, err := ioutil.ReadDir(SetCrdt1.GetDag().Nodes_storage_enplacement + "/remote")
			if err != nil {
				fmt.Printf("CheckUpdate - Checkupdate could not open folder\nerror: %s\n", err)
			}
			if len(files) >= 2*20 {
				time.Sleep(time.Duration(cfg.WaitTime) * time.Microsecond)
				strList = SetCrdt1.CheckUpdate_20Version(sema)
				fileRead.WriteString("CheckUpdates 20 Verison ! assure\n= = = = = = =\n")
			}
		} else {
			time.Sleep(time.Duration(cfg.WaitTime) * time.Microsecond)
			strList = SetCrdt1.CheckUpdate(sema)
		}
		if len(strList) > 0 {
			fileRead.WriteString("Just Received some updates\n")
			t := strconv.Itoa(GetTime(cfg.NtpServ))

			for j := 0; j < len(strList); j++ {
				getSema(sema, sys1.Ctx)
				file.WriteString(strList[j].Cid + "," + t + "," + strconv.Itoa(strList[j].RetrievalAlone) + "," + strconv.Itoa(strList[j].CalculTime) + ",0,0," + strconv.Itoa(strList[j].Time_decrypt) + "," + strconv.Itoa(strList[j].RetrievalTotal) + "," + strconv.Itoa(strList[j].ArrivalTime) + "," + strconv.Itoa(strList[j].TimeLookupFolder) + "," + strconv.Itoa(strList[j].TimeRemoveFiles) + "\n")
				returnSema(sema)
				fileRead.WriteString("writing 1 line in time.csv\n")
			}
			fileRead.WriteString("all update received are handled\n= = = = = = =\n")
		}
	}

	// Send updates concurrently every 1 seconds
	go sendUpdates(cfg.UpdatesNB, &SetCrdt1, cfg.NtpServ, file, sys1.Cr.Id, sema, cfg)

	//regularly scan files if there is any new received updates
	fmt.Printf("Starting the Set, updating %d times\n", cfg.UpdatesNB)
	ti = time.Now()
	k := 0
	for k < cfg.UpdatesNB {
		strList = make([]Set.TimeTuple, 0)
		if cfg.TestMode {
			files, err := ioutil.ReadDir(SetCrdt1.GetDag().Nodes_storage_enplacement + "/remote")
			if err != nil {
				fmt.Printf("CheckUpdate - Checkupdate could not open folder\nerror: %s\n", err)
			}
			if len(files) >= 2*20 {
				time.Sleep(time.Duration(cfg.WaitTime) * time.Microsecond)
				strList = SetCrdt1.CheckUpdate_20Version(sema)
				fileRead.WriteString("CheckUpdates 20 Verison ! assure\n= = = = = = =\n")
			}
		} else {
			time.Sleep(time.Duration(cfg.WaitTime) * time.Microsecond)
			strList = SetCrdt1.CheckUpdate(sema)
		}

		if len(strList) > 0 {
			fileRead.WriteString("Just Received some updates\n")
			t := strconv.Itoa(GetTime(cfg.NtpServ))
			for j := 0; j < len(strList); j++ {
				getSema(sema, sys1.Ctx)
				file.WriteString(strList[j].Cid + "," + t + "," + strconv.Itoa(strList[j].RetrievalAlone) + "," + strconv.Itoa(strList[j].CalculTime) + ",0,0," + strconv.Itoa(strList[j].Time_decrypt) + "," + strconv.Itoa(strList[j].RetrievalTotal) + "," + strconv.Itoa(strList[j].ArrivalTime) + "," + strconv.Itoa(strList[j].TimeLookupFolder) + "," + strconv.Itoa(strList[j].TimeRemoveFiles) + "\n")
				returnSema(sema)
				fileRead.WriteString("writing 1 line\n")
			}
			fileRead.WriteString("all update received are handled\n= = = = = = =\n")
		}
	}
	if err := file.Close(); err != nil {
		panic(fmt.Errorf("Error closing file\nerror : %s", err))
	}

}

var letterRunes = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func sendUpdates(nbUpdates int, SetCrdt1 *Set.CRDTSetOpBasedDag, ntpServ string, file *os.File, netID string, sema *semaphore.Weighted, cfg Config.IM_CRDTConfig) {
	fileWrite, _ := os.OpenFile(SetCrdt1.GetCRDTManager().Nodes_storage_enplacement+"/time/FileWrite.log", os.O_CREATE|os.O_WRONLY, 0755)
	fileWrite.WriteString(fmt.Sprintf("Starting the Set, updating %d times\n", nbUpdates))
	defer func(fileWrite *os.File) {
		fileWrite.WriteString("WRITE - all updates are done\n")
		fileWrite.Close()
	}(fileWrite)
	ti := time.Now()
	k := 0
	// s := make([]byte, 1048577)
	// fileWrite.WriteString(fmt.Sprintf("String initialized to 1mB, it took %d seconds\n", int(time.Since(ti).Seconds())))
	// for k < 1048577 {
	// 	if k%1000 == 0 {
	// 		fileWrite.WriteString(fmt.Sprintf("String initialized to 1mB, it took %d seconds\n", int(time.Since(ti).Seconds())))
	// 	}
	// 	s[k] = letterRunes[k%len(letterRunes)]
	// 	k = k + 1
	// }
	str := ""
	// str = string(s)
	// fileWrite.WriteString(fmt.Sprintf("String initialized to 1mB, it took %d seconds\n", int(time.Since(ti).Seconds())))
	k = 0
	for k < nbUpdates {
		time.Sleep(100 * time.Microsecond)

		if time.Since(ti) >= time.Second*time.Duration(cfg.SyncTime) {
			getSema(sema, context.Background())
			fileWrite.WriteString("updating the data\n")
			encodedCid, times := SetCrdt1.Add(netID + "VALUE ADDED : " + str + strconv.Itoa(k))
			fileWrite.WriteString("updating the data - taking sema\n")
			fileWrite.WriteString("Semaphore tooken\n")
			file.WriteString(encodedCid + "," + strconv.Itoa(GetTime(ntpServ)) + "," + "0,0," + strconv.Itoa(times.Time_add) + "," + strconv.Itoa(times.Time_encrypt) + ",0,0,0,0,0\n")
			fileWrite.WriteString("returning Semaphore\n")
			fileWrite.WriteString("WRITE - 1 line added to time.csv\n")
			returnSema(sema)
			k++
			ti = time.Now()
		}

	}
}

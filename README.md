# CRONUS Project

This project has been made in LORIA (INRIA Nancy, France).

It was originally nammed IM-CRDT, The namme has then been modified to CRONUS, to fit better the novelty bringed recently in this project.

It aims to develop and test integration of Merkle-CRDT in IPFS.
The developpement does define simple mutable data types such as String Set or Counter, however this is made to be able to represent any kind of mutable data, as long as each update it is transcripted in CRDT's Payload respecting the SEC property.

Multiple  Folder can be found here, choose one to see what you want to do:

- `CRDT_IPFS` is the implementation of IM-CRDT and IPFS_Alone. It does uses liP2P PubSub mechanism and IPFS as a file sharing mechanism 
- `ScriptExperiment` stores the Script I use to run my exepriments on Grid5000 (https://www.grid5000.fr/w/Grid5000:Home)
- `Results_25_09_23` does present the lastly ran results comparing __IM-CRDT__ and __IPFS Alone__

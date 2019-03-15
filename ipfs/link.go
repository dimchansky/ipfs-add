package ipfs

// Link represents an IPFS Merkle DAG Link between Nodes.
type Link struct {
	// multihash of the target object
	Cid Cid `json:"Cid"`
	// utf string name. should be unique per object
	Name string `json:"Name"`
	// cumulative size of target object
	Size uint64 `json:"Size"`
}

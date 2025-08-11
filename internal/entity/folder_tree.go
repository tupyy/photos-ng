package entity

// FolderNode represents a node in the folder tree structure
type FolderNode struct {
	// Path is the relative path of this folder from the root
	Path string

	// MediaFiles contains the list of media file paths directly in this folder
	MediaFiles []string

	// Children contains the immediate child folders
	Children []*FolderNode

	// Parent points to the parent folder (nil for root)
	Parent *FolderNode
}

// NewFolderNode creates a new folder node with the given path
func NewFolderNode(path string) *FolderNode {
	return &FolderNode{
		Path:       path,
		MediaFiles: make([]string, 0),
		Children:   make([]*FolderNode, 0),
		Parent:     nil,
	}
}

// AddChild adds a child folder to this node
func (fn *FolderNode) AddChild(child *FolderNode) {
	child.Parent = fn
	fn.Children = append(fn.Children, child)
}

// AddMediaFile adds a media file to this folder
func (fn *FolderNode) AddMediaFile(filePath string) {
	fn.MediaFiles = append(fn.MediaFiles, filePath)
}

// FindChild finds a direct child by path, returns nil if not found
func (fn *FolderNode) FindChild(path string) *FolderNode {
	for _, child := range fn.Children {
		if child.Path == path {
			return child
		}
	}
	return nil
}

// GetAllNodes returns all nodes in the tree (depth-first traversal)
func (fn *FolderNode) GetAllNodes() []*FolderNode {
	nodes := []*FolderNode{fn}
	for _, child := range fn.Children {
		nodes = append(nodes, child.GetAllNodes()...)
	}
	return nodes
}

// GetTotalMediaCount returns the total number of media files in this node and all descendants
func (fn *FolderNode) GetTotalMediaCount() int {
	total := len(fn.MediaFiles)
	for _, child := range fn.Children {
		total += child.GetTotalMediaCount()
	}
	return total
}

// GetTotalFolderCount returns the total number of folders in this subtree (including this node)
func (fn *FolderNode) GetTotalFolderCount() int {
	total := 1 // Count this node
	for _, child := range fn.Children {
		total += child.GetTotalFolderCount()
	}
	return total
}

/*
When first starting the tool, ConstructTree is called
from the main function of fs.go.

This file has all the functionality the daemon needs
to create the file system structure (tree + files
+ folders).
*/

package main

import (
	"fmt"

	"github.com/hanwen/go-fuse/fs"
	"github.com/hanwen/go-fuse/fuse"
	"golang.org/x/net/context"
)

// Every node in the filesystem needs a separate inode number.
// This global variable should be ++ incremented after each usage.
var inoIterator uint64 = 2

// Adds a file inode in the virtual filesystem tree.
func AddFile(ctx context.Context, node *fs.Inode, fileName string, fullPath string, modified bool) *fs.Inode {
	drpFileNode := DrpFileNode{}
	drpFileNode.drpPath = fullPath
	drpFileNode.modified = modified
	newfile := node.NewInode(
		ctx, &drpFileNode, fs.StableAttr{Ino: inoIterator})
	node.AddChild(fileName, newfile, false)

	inoIterator++

	return newfile
}

// Adds a folder inode in the virtual filesystem tree.
func AddFolder(ctx context.Context, node *fs.Inode, folderName string) *fs.Inode {
	dir := node.NewInode(
		ctx, &DrpFileNode{
			Data: []byte("sample dir data"),
			Attr: fuse.Attr{
				Mode: 0777,
			},
		}, fs.StableAttr{Ino: inoIterator, Mode: fuse.S_IFDIR})
	node.AddChild(folderName, dir, false)
	inoIterator++

	return dir
}

// Constructs the tree from our dropbox :)
// Given an array of DrpPaths ( a DrpPath is any path that exists
// as a file or folder in dropbox ), it generates the VFS tree.
// The contents of the files are NOT downloaded when the program
// is run. They are downloaded on the fly later, on open calls.
func ConstructTreeFromDrpPaths(ctx context.Context, r *HelloRoot, structure []DrpPath) {
	var m map[string](*fs.Inode) = make(map[string](*fs.Inode))

	m[""] = &r.Inode

	fmt.Println("Constructing tree")
	for _, entry := range structure {
		fmt.Println("Processing : " + entry.path)

		var containingFolder = firstPartFromPath(entry.path) // "/dirA" -> ""
		var newNodeName = lastFolderFromPath(entry.path)     // 		-> "dirA"

		fmt.Printf("containing folder : %v, newNodeName : %v \n", containingFolder, newNodeName)

		var parentNode = m[containingFolder]
		var newNode *fs.Inode
		if entry.isFolder {
			newNode = AddFolder(ctx, parentNode, newNodeName)
		} else {
			newNode = AddFile(ctx, parentNode, newNodeName, entry.path, false)
		}

		m[containingFolder+"/"+newNodeName] = newNode

		fmt.Println("Mapped the newly created node in " + containingFolder + "/" + newNodeName)
	}
}

// Given the FUSE root node, it constructs the tree.
// getDropboxTreeStructure() makes a Dropbox API call
// to generate all the dropbox folder structure
// modeled as an array of DrpPaths.
func ConstructTree(ctx context.Context, r *HelloRoot) {
	ConstructTreeFromDrpPaths(ctx, r, getDropboxTreeStructure())
}

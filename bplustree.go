package bplustree

import (
	"fmt"
	"strconv"
)

const (
	BPLUS_TREE_LEAF     = 0
	BPLUS_TREE_NON_LEAF = 1
	BORROW_FROM_LEFT    = 0
	BORROW_FROM_RIGHT   = 1
)

const DEBUG bool = true

func assert(scalar bool, ln int) {
	if !scalar && DEBUG {
		panic("assert failed, line: " + strconv.Itoa(ln))
	}
}

func tern(cond bool, r1, r2 interface{}) interface{} {
	if cond {
		return r1
	}
	return r2
}

func key_binary_search(arr []int, length int, target int) int {
	low, high := -1, length
	for low+1 < high {
		mid := low + (high-low)/2
		if target > arr[mid] {
			low = mid
		} else {
			high = mid
		}
	}
	if high >= length || arr[high] != target {
		return -high - 1
	}
	return high
}

func non_leaf_new() *bplus_non_leaf {
	return &bplus_non_leaf{
		kind: BPLUS_TREE_NON_LEAF,
	}
}

func leaf_new() *bplus_leaf {
	return &bplus_leaf{
		kind: BPLUS_TREE_LEAF,
	}
}

func non_leaf_delete(node *bplus_non_leaf) {
	//free(node);
	node = nil
}

func leaf_delete(node *bplus_leaf) {
	//free(node);
	node = nil
}

func bplus_tree_search(tree *bplus_tree, key int) int {

	var node *bplus_node = tree.root

	for node != nil {
		switch (*node).getKind() {
		case BPLUS_TREE_NON_LEAF:
			nln := (*node).(*bplus_non_leaf)
			i := key_binary_search(nln.key[:], nln.children-1, key)
			if i >= 0 {
				node = nln.sub_ptr[i+1]
			} else {
				i = -i - 1
				node = nln.sub_ptr[i]
			}
		case BPLUS_TREE_LEAF:
			ln := (*node).(*bplus_leaf)
			i := key_binary_search(ln.key[:], ln.entries, key)
			if i >= 0 {
				return ln.data[i]
			}
			return 0 // NOTE: what does a zero return value indicate??
		default:
			assert(false, 80)
			//log.Println("bplustree.go bplus_tree_search hit default")
		}
	}
	return 0 // NOTE: what does a zero return value indicate??
}

func non_leaf_insert(tree *bplus_tree, node *bplus_non_leaf, sub_node *bplus_node, key int, level int) int {

	var i, j, split_key int
	var split int = 0
	var sibling *bplus_non_leaf

	var insert int = key_binary_search(node.key[:], node.children-1, key)
	assert(insert < 0, 95)
	insert = -insert - 1

	/* node full */
	if node.children == tree.order {
		/* split = [m/2] */
		split = (tree.order + 1) / 2
		/* splited sibling node */
		sibling = non_leaf_new()
		sibling.next = node.next
		node.next = sibling
		/* non-leaf node's children always equals to split + 1 after insertion */
		node.children = split + 1
		/* sibling node replication due to location of insertion */
		if insert < split {
			split_key = node.key[split-1]
			/* sibling node's first sub-node */
			sibling.sub_ptr[0] = node.sub_ptr[split]
			node.sub_ptr[split].parent = sibling
			/* insertion point is before split point, replicate from key[split] */
			for i, j = split, 0; i < tree.order-1; i, j = i+1, j+1 {
				sibling.key[j] = node.key[i]
				sibling.sub_ptr[j+1] = node.sub_ptr[i+1]
				node.sub_ptr[i+1].parent = sibling
			}
			sibling.children = j + 1
			/* insert new key and sub-node */
			for i = node.children - 2; i > insert; i-- {
				node.key[i] = node.key[i-1]
				node.sub_ptr[i+1] = node.sub_ptr[i]
			}
			node.key[i] = key
			node.sub_ptr[i+1] = sub_node
			sub_node.parent = node
		} else if insert == split {
			split_key = key
			/* sibling node's first sub-node */
			sibling.sub_ptr[0] = sub_node
			sub_node.parent = sibling
			/* insertion point is split point, replicate from key[split] */
			for i, j = split, 0; i < tree.order-1; i, j = i+1, j+1 {
				sibling.key[j] = node.key[i]
				sibling.sub_ptr[j+1] = node.sub_ptr[i+1]
				node.sub_ptr[i+1].parent = sibling
			}
			sibling.children = j + 1
		} else {
			split_key = node.key[split]
			/* sibling node's first sub-node */
			sibling.sub_ptr[0] = node.sub_ptr[split+1]
			node.sub_ptr[split+1].parent = sibling
			/* insertion point is after split point, replicate from key[split + 1] */
			for i, j = split+1, 0; i < tree.order-1; j++ {
				if j != insert-split-1 {
					sibling.key[j] = node.key[i]
					sibling.sub_ptr[j+1] = node.sub_ptr[i+1]
					node.sub_ptr[i+1].parent = sibling
					i++
				}
			}
			/* reserve a hole for insertion */
			if j > insert-split-1 {
				sibling.children = j + 1
			} else {
				assert(j == insert-split-1, 159)
				sibling.children = j + 2
			}
			/* insert new key and sub-node*/
			j = insert - split - 1
			sibling.key[j] = key
			sibling.sub_ptr[j+1] = sub_node
			sub_node.parent = sibling
		}
	} else {
		/* simple insertion */
		for i = node.children - 1; i > insert; i-- {
			node.key[i] = node.key[i-1]
			node.sub_ptr[i+1] = node.sub_ptr[i]
		}
		node.key[i] = key
		node.sub_ptr[i+1] = sub_node
		node.children++
	}
	if split { // NOTE: split is an int; in C the int's 0 and 1 can also be looked at as booleans
		var parent *bplus_non_leaf = node.parent
		if parent == nil {
			// used to be ++level, changed it to level+1
			if level+1 >= tree.level {
				panic("!!Level exceeded, please expand the tree level, non-leaf order or leaf entries for element capacity!\n")
				node.next = sibling.next
				non_leaf_delete(sibling)
				return -1 // NOTE: may not need??
			}
			/* new parent */
			parent = non_leaf_new()
			parent.key[0] = split_key
			parent.sub_ptr[0] = node.(*bplus_node)
			parent.sub_ptr[1] = sibling.(*bplus_node)
			parent.children = 2
			/* update root */
			tree.root = parent.(*bplus_node)
			tree.head[level] = parent.(*bplus_node)
			node.parent = parent
			sibling.parent = parent
		} else {
			/* Trace upwards */
			sibling.parent = parent
			return non_leaf_insert(tree, parent, sibling.(*bplus_node), split_key, level+1)
		}
	}
	return 0 // NOTE: what is this return value even used for??
}

func leaf_insert(tree *bplus_tree, leaf *bplus_leaf, key int, data int) int {

	var i, j, split int = 0
	var sibling *bplus_leaf

	var insert int = key_binary_search(leaf.key, leaf.entries, key)
	if insert >= 0 {
		/* Already exists */
		return -1 // NOTE: ??
	}
	insert = -insert - 1

	/* node full */
	if leaf.entries == tree.entries {
		/* split = [m/2] */
		split = (tree.entries + 1) / 2
		/* splited sibling node */
		sibling = leaf_new()
		sibling.next = leaf.next
		leaf.next = sibling
		/* leaf node's entries always equals to split after insertion */
		leaf.entries = split
		/* sibling leaf replication due to location of insertion */
		if insert < split {
			/* insertion point is before split point, replicate from key[split - 1] */
			for i, j = split-1, 0; i < tree.entries; i, j = i+1, j+1 {
				sibling.key[j] = leaf.key[i]
				sibling.data[j] = leaf.data[i]
			}
			sibling.entries = j
			/* insert new key and sub-node */
			for i = split - 1; i > insert; i-- {
				leaf.key[i] = leaf.key[i-1]
				leaf.data[i] = leaf.data[i-1]
			}
			leaf.key[i] = key
			leaf.data[i] = data
		} else {
			/* insertion point is or after split point, replicate from key[split] */
			for i, j = split, 0; i < tree.entries; j++ {
				if j != insert-split {
					sibling.key[j] = leaf.key[i]
					sibling.data[j] = leaf.data[i]
					i++
				}
			}
			/* reserve a hole for insertion */
			if j > insert-split {
				sibling.entries = j
			} else {
				assert(j == insert-split, 259)
				sibling.entries = j + 1
			}
			/* insert new key */
			j = insert - split
			sibling.key[j] = key
			sibling.data[j] = data
		}
	} else {
		/* simple insertion */
		for i = leaf.entries; i > insert; i-- {
			leaf.key[i] = leaf.key[i-1]
			leaf.data[i] = leaf.data[i-1]
		}
		leaf.key[i] = key
		leaf.data[i] = data
		leaf.entries++
	}

	if split {
		var parent *bplus_non_leaf = leaf.parent
		if parent == nil {
			/* new parent */
			parent = non_leaf_new()
			parent.key[0] = sibling.key[0]
			parent.sub_ptr[0] = leaf.(*bplus_node)
			parent.sub_ptr[1] = sibling.(*bplus_node)
			parent.children = 2
			/* update root */
			tree.root = parent.(*bplus_node)
			tree.head[1] = parent.(*bplus_node)
			leaf.parent = parent
			sibling.parent = parent
		} else {
			/* trace upwards */
			sibling.parent = parent
			return non_leaf_insert(tree, parent, sibling.(*bplus_node), sibling.key[0], 1) // NOTE: what does this return??
		}
	}
	return 0 //NOTE: ??
}

func bplus_tree_insert(tree *bplus_tree, key int, data int) int {

	var node *bplus_node = tree.root

	for node != nil {
		switch node.getKind() {
		case BPLUS_TREE_NON_LEAF:
			nln := node.(*bplus_non_leaf)
			i := key_binary_search(nln.key, nln.children-1, key)
			if i >= 0 {
				node = nln.sub_ptr[i+1]
			} else {
				i = -i - 1
				node = nln.sub_ptr[i]
			}
		case BPLUS_TREE_LEAF:
			ln := node.(*bplus_leaf)
			return leaf_insert(tree, ln, key, data)
		default:
			assert(false, 320)
			//log.Println("bplustree.go bplus_tree_insert hit default")
		}
	}

	/* new root */
	root := leaf_new()
	root.key[0] = key
	root.data[0] = data
	root.entries = 1

	tree.head[0] = root.(*bplus_node)
	tree.root = root.(*bplus_node)
	return 0
}

func non_leaf_remove(tree *bplus_tree, node *bplus_non_leaf, remove int, level int) {

	var i, j, k int
	var sibling *bplus_non_leaf

	if node.children <= (tree.order+1)/2 {
		var parent *bplus_non_leaf = node.parent
		if parent != nil {
			var borrow int = 0
			/* find which sibling node with same parent to be borrowed from */
			i = key_binary_search(parent.key, parent.children-1, node.key[0])
			assert((i < 0), 346)
			i = -i - 1
			if i == 0 {
				/* no left sibling, choose right one */
				sibling = parent.sub_ptr[i+1].(*bplus_non_leaf)
				borrow = BORROW_FROM_RIGHT
			} else if i == parent.children-1 {
				/* no right sibling, choose left one */
				sibling = parent.sub_prt[i-1].(*bplus_non_leaf)
				borrow = BORROW_FROM_LEFT
			} else {
				var l_sib *bplus_non_leaf = parent.sub_ptr[i-1].(*bplus_non_leaf)
				var r_sib *bplus_non_leaf = parent.sub_ptr[i+1].(*bplus_non_leaf)
				/* if both left and right sibling found, choose the one with more children */

				// NOTE: using a home grown ternary function here to compensate for the fact that go
				// does not have one. don't know at current time if this will result in poor performance.
				sibling = tern(l_sib.children >= r.sib.children, l_sib, r_sib)
				borrow = tern(l_sib.children >= r.sib.children, BORROW_FROM_LEFT, BORROW_FROM_RIGHT)
			}

			/* locate parent node key to update later */
			i = i - 1

			if borrow == BORROW_FROM_LEFT {
				if sibling.children > (tree.order+1)/2 {
					/* node's elements right shift */
					for j = remove; j > 0; j-- {
						node.key[j] = node.key[j-1]
					}
					for j = remove + 1; j > 0; j-- {
						node.sub_ptr[j] = node.sub_ptr[j-1]
					}
					/* parent key right rotation */
					node.key[0] = parent.key[i]
					parent.key[i] = sibling.key[sibling.children-2]
					/* borrow the last sub-node from left sibling */
					node.sub_ptr[0] = sibling.sub_ptr[sibling.children-1]
					sibling.sub_ptr[sibling.children-1].parent = node
					sibling.children--
				} else {
					/* move parent key down */
					sibling.key[sibling.children-1] = parent.key[i]
					/* merge with left sibling */
					for j, k = sibling.children, 0; k < node.children-1; k++ {
						if k != remove {
							sibling.key[j] = node.key[k]
							j++
						}
					}
					for j, k = sibling.children, 0; k < node.children; k++ {
						if k != remove+1 {
							sibling.sub_ptr[j] = node.sub_ptr[k]
							node.sub_ptr[k].parent = sibling
							j++
						}
					}
					sibling.children = j
					/* delete merged node */
					sibling.next = node.next
					non_leaf_delete(node)
					/* trace upwards */
					non_leaf_remove(tree, parent, i, level+1)
				}
			} else {
				/* remove key first in case of overflow during merging with sibling node */
				for remove < node.children-2 {
					node.key[remove] = node.key[remove+1]
					node.sub_ptr[remove+1] = node.sub_ptr[remove+2]
					remove++
				}
				node.children--
				if sibling.children > (tree.order+1)/2 {
					/* parent key left rotation */
					node.key[node.children-1] = parent.key[i+1]
					parent.key[i+1] = sibling.key[0]
					/* borrow the frist sub-node from right sibling */
					node.sub_ptr[node.children] = sibling.sub_ptr[0]
					sibling.sub_ptr[0].parent = node
					node.children++
					/* left shift in right sibling */
					for j = 0; j < sibling.children-2; j++ {
						sibling.key[j] = sibling.key[j+1]
					}
					for j = 0; j < sibling.children-1; j++ {
						sibling.sub_ptr[j] = sibling.sub_ptr[j+1]
					}
					sibling.children--
				} else {
					/* move parent key down */
					node.key[node.children-1] = parent.key[i+1]
					node.children++
					/* merge with right sibling */
					for j, k = node.children-1, 0; k < sibling.children-1; j, k = j+1, k+1 {
						node.key[j] = sibling.key[k]
					}
					for j, k = node.children-1, 0; k < sibling.children; j, k = j+1, k+1 {
						node.sub_ptr[j] = sibling.sub_ptr[k]
						sibling.sub_ptr[k].parent = node
					}
					node.children = j
					/* delete merged sibling */
					node.next = sibling.next
					non_leaf_delete(sibling)
					/* trace upwards */
					non_leaf_remove(tree, parent, i+1, level+1)
				}
			}
			/* deletion finishes */
			return // NOTE: not sure what this means in the long and short??
		} else {
			if node.children == 2 {
				/* delete old root node */
				assert(remove == 0, 467)
				node.sub_ptr[0].parent = nil
				tree.root = node.sub_ptr[0]
				tree.head[level] = nil
				non_leaf_delete(node)
				return // NOTE: again, not sure about this empty return... or what it means for go translation...
			}
		}
	}

	/* simple deletion */
	assert(node.children > 2, 478)
	for remove < node.children-2 {
		node.key[remove] = node.key[remove+1]
		node.sub_ptr[remove+1] = node.sub_ptr[remove+2]
		remove++
	}
	node.children--
}

func leaf_remove(tree *bplus_tree, leaf *bplus_leaf, key int) int {

	var i, j, k int
	var sibling *bplus_leaf

	var remove int = key_binary_search(leaf.key, leaf.entries, key)
	if remove < 0 {
		/* Not exist */
		return -1 // NOTE: whatever this means...
	}

	if leaf.entries <= (tree.entries+1)/2 {
		var parent *bplus_non_leaf = leaf.parent
		if parent != nil {
			var borrow int = 0
			/* find which sibling node with same parent to be borrowed from */
			i = key_binary_search(parent.key, parent.children-1, leaf.key[0])
			if i >= 0 {
				i = i + 1
				if i == parent.children-1 {
					/* the last node, no right sibling, choose left one */
					sibling = parent.sub_ptr[i-1].(*bplus_leaf)
					borrow = BORROW_FROM_LEFT
				} else {
					var l_sib *bplus_leaf = parent.sub_ptr[i-1].(*bplus_leaf)
					var r_sib *bplus_leaf = parent.sub_ptr[i+1].(*bplus_leaf)
					/* if both left and right sibling found, choose the one with more entries */

					// NOTE: again, used my home grown ternary function, don't knwo if this is good or
					// bad, just mentioning it in case it it bad, then this portion can be re-written...
					sibling = tern(l_sib.entries >= r_sib.entries, l_sib, r_sib)
					borrow = tern(l_sib.entries >= r_sib.entries, BORROW_FROM_LEFT, BORROW_FROM_RIGHT)
				}
			} else {
				i = -i - 1
				if i == 0 {
					/* the frist node, no left sibling, choose right one */
					sibling = parent.sub_ptr[i+1].(*bplus_leaf)
					borrow = BORROW_FROM_RIGHT
				} else if i == parent.children-1 {
					/* the last node, no right sibling, choose left one */
					sibling = parent.sub_ptr[i-1].(*bplus_leaf)
					borrow = BORROW_FROM_LEFT
				} else {
					var l_sib *bplus_leaf = parent.sub_ptr[i-1].(*bplus_leaf)
					var r_sib *bplus_leaf = parent.sub_ptr[i+1].(*bplus_leaf)
					/* if both left and right sibling found, choose the one with more entries */

					// NOTE: again, used my home grown ternary function, don't knwo if this is good or
					// bad, just mentioning it in case it it bad, then this portion can be re-written...
					sibling = tern(l_sib.entries >= r_sib.entries, l_sib, r_sib)
					borrow = tern(l_sib.entries >= r_sib.entries, BORROW_FROM_LEFT, BORROW_FROM_RIGHT)

				}
			}

			/* locate parent node key to update later */
			i = i - 1

			if borrow == BORROW_FROM_LEFT {
				if sibling.entries > (tree.entries+1)/2 {
					/* right shift in leaf node */
					for remove > 0 {
						leaf.key[remove] = leaf.key[remove-1]
						leaf.data[remove] = leaf.data[remove-1]
						remove--
					}
					/* borrow the last element from left sibling */
					leaf.key[0] = sibling.key[sibling.entries-1]
					leaf.data[0] = sibling.data[sibling.entries-1]
					sibling.entries--
					/* update parent key */
					parent.key[i] = leaf.key[0]
				} else {
					/* merge with left sibling */
					for j, k = sibling.entries, 0; k < leaf.entries; k++ {
						if k != remove {
							sibling.key[j] = leaf.key[k]
							sibling.data[j] = leaf.data[k]
							j++
						}
					}
					sibling.entries = j
					/* delete merged leaf */
					sibling.next = leaf.next
					leaf_delete(leaf)
					/* trace upwards */
					non_leaf_remove(tree, parent, i, 1)
				}
			} else {
				/* remove element first in case of overflow during merging with sibling node */
				for remove < leaf.entries-1 {
					leaf.key[remove] = leaf.key[remove+1]
					leaf.data[remove] = leaf.data[remove+1]
					remove++
				}
				leaf.entries--
				if sibling.entries > (tree.entries+1)/2 {
					/* borrow the first element from right sibling */
					leaf.key[leaf.entries] = sibling.key[0]
					leaf.data[leaf.entries] = sibling.data[0]
					leaf.entries++
					/* left shift in right sibling */
					for j = 0; j < sibling.entries-1; j++ {
						sibling.key[j] = sibling.key[j+1]
						sibling.data[j] = sibling.data[j+1]
					}
					sibling.entries--
					/* update parent key */
					parent.key[i+1] = sibling.key[0]
				} else {
					/* merge with right sibling */
					for j, k = leaf.entries, 0; k < sibling.entries; j, k = j+1, k+1 {
						leaf.key[j] = sibling.key[k]
						leaf.data[j] = sibling.data[k]
					}
					leaf.entries = j
					/* delete right sibling */
					leaf.next = sibling.next
					leaf_delete(sibling)
					/* trace upwards */
					non_leaf_remove(tree, parent, i+1, 1)
				}
			}
			/* deletion finishes */
			return 0
		} else {
			if leaf.entries == 1 {
				/* delete the only last node */
				assert(key == leaf.key[0])
				tree.root = nil
				tree.head[0] = nil
				leaf_delete(leaf)
				return 0
			}
		}
	}

	/* simple deletion */
	for remove < leaf.entries-1 {
		leaf.key[remove] = leaf.key[remove+1]
		leaf.data[remove] = leaf.data[remove+1]
		remove++
	}
	leaf.entries--

	return 0
}

func bplus_tree_delete(tree *bplus_tree, key int) int {

	var node *bplus_node = tree.root

	for node != nil {
		switch node.kind {
		case BPLUS_TREE_NON_LEAF:
			nln := node.(*bplus_non_leaf)
			i := key_binary_search(nln.key, nln.children-1, key)
			if i >= 0 {
				node = nln.sub_ptr[i+1]
			} else {
				i = -i - 1
				node = nln.sub_ptr[i]
			}
		case BPLUS_TREE_LEAF:
			ln := node.(*bplus_leaf)
			return leaf_remove(tree, ln, key)
		default:
			// was assert(0)
			assert(false, 659)
		}
	}

	return -1 // NOTE: again, don't know what this return value actually signifies
}

func bplus_tree_dump(tree *bplus_tree) {

	var i, j int

	for i = tree.level - 1; i > 0; i-- {
		var node *bplus_non_leaf = tree.head[i].(*bplus_non_leaf)
		if node != nil {
			fmt.Printf("LEVEL %d:\n", i)
			for node != nil {
				fmt.Printf("node: ")
				for j = 0; j < node.children-1; j++ {
					fmt.Printf("%d ", node.key[j])
				}
				fmt.Printf("\n")
				node = node.next
			}
		}
	}

	var leaf *bplus_leaf = tree.head[0].(*bplus_leaf)
	if leaf != nil {
		fmt.Printf("LEVEL 0:\n")
		for leaf != nil {
			fmt.Printf("leaf: ")
			for j = 0; j < leaf.entries; j++ {
				fmt.Printf("%d ", leaf.key[j])
			}
			printf("\n")
			leaf = leaf.next
		}
	} else {
		printf("Empty tree!\n")
	}
}

func bplus_tree_get(tree *bplus_tree, key int) int {
	var data int = bplus_tree_search(tree, key)
	if data >= 0 {
		return data
	}
	return -1
}

func bplus_tree_put(tree *bplus_tree, key int, data int) int {
	if data >= 0 {
		return bplus_tree_insert(tree, key, data)
	}
	return bplus_tree_delete(tree, key)
}

func bplus_tree_init(level int, order int, entries int) *bplus_tree {
	/* The max order of non leaf nodes must be more than two */
	assert(MAX_ORDER > MIN_ORDER, 715)
	assert(level <= MAX_LEVEL && order <= MAX_ORDER && entries <= MAX_ENTRIES, 716)

	var tree *bplus_tree = make(*tree, 0)
	if tree != nil {
		tree.root = nil
		tree.level = level
		tree.order = order
		tree.entries = entries
		//memset(tree.head, 0, MAX_LEVEL * sizeof(struct bplus_node *));
	}

	return tree
}

func bplus_tree_deinit(tree *bplus_tree) {
	//free(tree);
	tree = nil
}

func bplus_tree_get_range(tree *bplus_tree, key1 int, key2 int) int {

	var data, min, max int = 0

	if key1 <= key2 {
		min = key1
	} else {
		min = key2
	}

	if min == key1 {
		max = key2
	} else {
		max = key1
	}

	var node *bplus_node = tree.root

	for node != nil {
		switch node.getKind() {
		case BPLUS_TREE_NON_LEAF:
			nln := node.(*bplus_non_leaf)
			i := key_binary_search(nln.key, nln.children-1, min)
			if i >= 0 {
				node = nln.sub_ptr[i+1]
			} else {
				i = -i - 1
				node = nln.sub_ptr[i]
			}
		case BPLUS_TREE_LEAF:
			ln := node.(*bplus_leaf)
			i := key_binary_search(ln.key, ln.entries, min)
			if i < 0 {
				i = -i - 1
				if i >= ln.entries {
					ln = ln.next
				}
			}
			for ln != nil && ln.key[i] <= max {
				data = ln.data[i]
				if i+1 >= ln.entries {
					ln = ln.next
					i = 0
				}
			}
			return data
		default:
			//assert(0)
			assert(false, 784)
		}
	}
	return 0
}

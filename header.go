package bplustree

const MIN_ORDER = 3
const MAX_ORDER = 64
const MAX_ENTRIES = 64
const MAX_LEVEL = 10

type bplus_node interface {
	getKind() int
	getParent() *bplus_non_leaf
}

type bplus_non_leaf struct {
	kind     int
	parent   *bplus_non_leaf
	next     *bplus_non_leaf
	children int
	key      [MAX_ORDER - 1]int
	sub_ptr  [MAX_ORDER]*bplus_node
}

type bplus_leaf struct {
	kind    int
	parent  *bplus_non_leaf
	next    *bplus_leaf
	entries int
	key     [MAX_ENTRIES]int
	data    [MAX_ENTRIES]int
}

type bplus_tree struct {
	order   int
	entries int
	level   int
	root    *bplus_node
	head    [MAX_LEVEL]*bplus_node
}

type bplus_tree_dump func(tree *bplus_tree)
type bplus_tree_get func(tree *bplus_tree, key int) int
type bplus_tree_put func(tree *bplus_tree, key, data int) int
type bplus_tree_get_range func(tree *bplus_tree, key1, key2 int) int
type bplus_tree_init func(level, order, entries int) *bplus_tree
type bplus_tree_deinit func(tree *bplus_tree)

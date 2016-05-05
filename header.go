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

func (nln *bplus_non_leaf) getKind() int {
	return nln.kind
}

func (nln *bplus_non_leaf) getParent() *bplus_non_leaf {
	return nln.parent
}

type bplus_leaf struct {
	kind    int
	parent  *bplus_non_leaf
	next    *bplus_leaf
	entries int
	key     [MAX_ENTRIES]int
	data    [MAX_ENTRIES]int
}

func (ln *bplus_leaf) getKind() int {
	return ln.kind
}

func (nln *bplus_leaf) getParent() *bplus_non_leaf {
	return ln.parent
}

type bplus_tree struct {
	order   int
	entries int
	level   int
	root    *bplus_node
	head    [MAX_LEVEL]*bplus_node
}

type btree interface {
	bplus_tree_dump(tree *bplus_tree)
	bplus_tree_get(tree *bplus_tree, key int) int
	bplus_tree_put(tree *bplus_tree, key, data int) int
	bplus_tree_get_range(tree *bplus_tree, key1, key2 int) int
	bplus_tree_init(level, order, entries int) *bplus_tree
	bplus_tree_deinit(tree *bplus_tree)
}

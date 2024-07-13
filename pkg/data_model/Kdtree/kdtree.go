package Kdtree

import (
	"fmt"
	"sort"
	"strings"

	"github.com/PES-Innovation-Lab/willow-go/pkg/data_model/Kdtree/queue"

	kdtree "github.com/PES-Innovation-Lab/willow-go/pkg/data_model/Kdtree/KDTreeEncoding"

	flatbuffers "github.com/google/flatbuffers/go"
)

type Relation int

const (
	Lesser Relation = iota
	Equal
	Greater
)

var ErrTreeNotSetup = fmt.Errorf("tree is not setup, make sure you create the tree using NewTree")

type Comparable[T any] interface {
	fmt.Stringer
	Order(rhs T, dim int) Relation
	DistDim(rhs T, dim int) int
	Dist(rhs T) int
	Encode() []byte
}

type KDTree[T Comparable[T]] struct {
	Dimensions int
	Root       *KdNode[T]
	IsSetup    bool
	ZeroVal    T
	Sz         int
}

type KdNode[T Comparable[T]] struct {
	Value T
	Left  *KdNode[T]
	Right *KdNode[T]
}

func NewKDNode[T Comparable[T]](value T) *KdNode[T] {
	return &KdNode[T]{
		Value: value,
	}
}

func NewKDTreeWithValues[T Comparable[T]](d int, vs []T) *KDTree[T] {
	sz := len(vs)
	initialIndices := make([][]int, d)
	for cd := range initialIndices {
		initialIndices[cd] = iotaSlice(len(vs))
		sort.Slice(initialIndices[cd], func(i, j int) bool {
			return vs[initialIndices[cd][i]].Order(vs[initialIndices[cd][j]], cd) == Lesser
		})
	}
	root := insertAllNew[T](vs, initialIndices, 0)
	// root := insertAllOld(d, vs, 0)
	return &KDTree[T]{
		Dimensions: d,
		Root:       root,
		IsSetup:    true,
		Sz:         sz,
	}
}

func (t *KDTree[T]) FindMin(targetDimension int) (T, bool) {
	if t.Root == nil || targetDimension >= t.Dimensions {
		return t.ZeroVal, false
	}
	res := findMin(t.Dimensions, targetDimension, 0, t.Root)
	if res == nil {
		return t.ZeroVal, false
	}
	return *res, true
}

func (t *KDTree[T]) NearestNeighbor(value T) (T, bool) {
	res := nearestNeighbor(t.Dimensions, &value, nil, 0, t.Root)
	if res == nil {
		return t.ZeroVal, false
	}
	return *res, true
}

func (t *KDTree[T]) Add(value T) bool {
	if t.Root == nil {
		t.Root = NewKDNode(value)
		return true
	}
	res := add(t.Dimensions, value, 0, t.Root)
	if res {
		t.Sz++
	}
	return res
}

func (t *KDTree[T]) Delete(value T) bool {
	ok := false
	t.Root, ok = deleteNode(t.Dimensions, value, 0, t.Root)
	if ok {
		t.Sz--
	}
	return ok
}

func (t *KDTree[T]) String() string {
	b := strings.Builder{}
	var q queue.Queue[*KdNode[T]] = queue.NewLLQueue[*KdNode[T]]()
	q.Push(t.Root)
	for !q.Empty() {
		sz := q.Size()
		for i := 0; i < sz; i++ {
			n, _ := q.Pop()
			if n != nil {
				b.WriteString(n.Value.String())
				b.WriteString(", ")
				q.Push(n.Left)
				q.Push(n.Right)
			} else {
				b.WriteString("nil, ")
			}
		}
		b.WriteString("\n")
	}
	return b.String()
}

const encodingVersion uint32 = 0

func (t *KDTree[T]) Encode() []byte {
	encodedPreorderItems := preorderTraversal(t.Root)
	itemCount := len(encodedPreorderItems)
	if itemCount != t.Sz {
		msg := fmt.Sprintf("itemCount (%d) and t.sz (%d) don't have the same size! Some bookkeeping has gone wrong!", itemCount, t.Sz)
		panic(msg)
	}
	encodedInorderIndices := inorderTraversal(t.Root, t.Sz)

	builder := flatbuffers.NewBuilder(256)

	kdtree.KDTreeStartInorderIndicesVector(builder, itemCount)
	for i := itemCount - 1; i >= 0; i-- {
		idx := encodedInorderIndices[i]
		builder.PrependInt64(int64(idx))
	}
	inorderIndices := builder.EndVector(itemCount)

	var encodedItems []flatbuffers.UOffsetT
	for i := 0; i < itemCount; i++ {
		item := encodedPreorderItems[i]
		sz := len(item)

		kdtree.ItemStartDataVector(builder, sz)
		for i := sz - 1; i >= 0; i-- {
			itemByte := item[i]
			builder.PrependByte(itemByte)
		}
		itemBytesVector := builder.EndVector(sz)

		kdtree.ItemStart(builder)
		kdtree.ItemAddData(builder, itemBytesVector)
		encodedItem := kdtree.ItemEnd(builder)

		encodedItems = append(encodedItems, encodedItem)
	}

	kdtree.KDTreeStartItemsVector(builder, itemCount)
	for i := itemCount - 1; i >= 0; i-- {
		builder.PrependUOffsetT(encodedItems[i])
	}
	items := builder.EndVector(itemCount)

	kdtree.KDTreeStart(builder)
	kdtree.KDTreeAddVersionNumber(builder, encodingVersion)
	kdtree.KDTreeAddDimensions(builder, uint32(t.Dimensions))
	kdtree.KDTreeAddInorderIndices(builder, inorderIndices)
	kdtree.KDTreeAddItems(builder, items)
	encodedKDTree := kdtree.KDTreeEnd(builder)
	builder.Finish(encodedKDTree)
	return builder.FinishedBytes()
}

func NewKDTreeFromBytes[T Comparable[T]](encodedBytes []byte, decodeItemFunc func([]byte) T) *KDTree[T] {
	tree := kdtree.GetRootAsKDTree(encodedBytes, 0)
	if encodingVersion != tree.VersionNumber() {
		panic("Unsupported encoding version number!")
	}
	itemsLength := tree.ItemsLength()
	if itemsLength != tree.InorderIndicesLength() {
		msg := fmt.Sprintf("The number of the indices (%d) are not the same as the number of items(%d)!",
			tree.InorderIndicesLength(), itemsLength)
		panic(msg)
	}
	// Note: This will be useful when I need to reconstruct the exact tree again.
	// For now the reconstructed tree will not be exactly the same. It will be a rebalanced tree.
	// This should not introduce any bugs in the code but rather make the tree based operations faster, when
	// loaded from binary.
	// preorderIndices := iotaSlice(itemsLength)
	// inorderIndices := make([]int, itemsLength)
	// inorderIndexLookup := make([]int, itemsLength)
	// for i := range inorderIndices {
	// 	idx := int(tree.InorderIndices(i))
	// 	inorderIndices[i] = idx
	// 	inorderIndexLookup[idx] = i
	// }
	items := make([]T, itemsLength)
	for i := 0; i < itemsLength; i++ {
		itemPtr := new(kdtree.Item)
		if tree.Items(itemPtr, i) {
			item := decodeItemFunc(itemPtr.DataBytes())
			items[i] = item
		}
	}
	dimensions := int(tree.Dimensions())
	return NewKDTreeWithValues(dimensions, items)
}

func preorderTraversal[T Comparable[T]](r *KdNode[T]) [][]byte {
	var res [][]byte
	preorderTraversalImpl(r, &res)
	return res
}

func preorderTraversalImpl[T Comparable[T]](r *KdNode[T], res *[][]byte) {
	if r == nil {
		return
	}
	*res = append(*res, r.Value.Encode())
	preorderTraversalImpl(r.Left, res)
	preorderTraversalImpl(r.Right, res)
}

func inorderTraversal[T Comparable[T]](r *KdNode[T], sz int) []int {
	preorderIndex := 0
	inorderIndex := 0
	res := make([]int, sz)
	inorderTraversalImpl(r, &preorderIndex, &inorderIndex, &res)
	return res
}

func inorderTraversalImpl[T Comparable[T]](r *KdNode[T], preorderIndex, inorderIndex *int, res *[]int) {
	if r == nil {
		return
	}
	currPreorderIndex := *preorderIndex
	*preorderIndex++
	inorderTraversalImpl(r.Left, preorderIndex, inorderIndex, res)
	currInorderIndex := *inorderIndex
	*inorderIndex++
	(*res)[currInorderIndex] = currPreorderIndex
	inorderTraversalImpl(r.Right, preorderIndex, inorderIndex, res)
}

func insertAllNew[T Comparable[T]](vs []T, initialIndices [][]int, cd int) *KdNode[T] {
	if len(initialIndices[0]) == 0 {
		return nil
	}
	dims := len(initialIndices)
	cutIndex := initialIndices[0]
	mv, mvIdx := midValue(vs, cutIndex)
	n := NewKDNode(mv)

	// Split initialIndices
	temp := make([]int, len(cutIndex))
	copy(temp, cutIndex)

	lh := make([][]int, dims)
	uh := make([][]int, dims)
	si := (len(initialIndices[0]) - 1) / 2
	for i := 0; i < dims; i++ {
		indexArray := initialIndices[i]
		lh[i] = indexArray[:si]
		uh[i] = indexArray[si+1:]
	}

	for i := 1; i < dims; i++ {
		lhi := 0
		uhi := 0
		indexArray := initialIndices[i]
		for _, idx := range indexArray {
			if idx == mvIdx {
				continue
			}
			v := vs[idx]
			if v.Order(mv, cd) == Lesser {
				lh[i-1][lhi] = idx
				lhi++
			} else {
				uh[i-1][uhi] = idx
				uhi++
			}
		}
	}
	copy(initialIndices[dims-1], temp)

	ncd := (cd + 1) % dims
	n.Left = insertAllNew(vs, lh, ncd)
	n.Right = insertAllNew(vs, uh, ncd)
	return n
}

func midValue[T Comparable[T]](vs []T, cutIndex []int) (T, int) {
	i := (len(cutIndex) - 1) / 2
	mvi := cutIndex[i]
	return vs[mvi], mvi
}

func iotaSlice(n int) []int {
	s := make([]int, n)
	for i := range s {
		s[i] = i
	}
	return s
}

func deleteNode[T Comparable[T]](d int, value T, cd int, r *KdNode[T]) (*KdNode[T], bool) {
	if r == nil {
		return nil, false
	}
	ncd := (cd + 1) % d
	ok := false
	if r.Value.Dist(value) == 0 {
		ok = true
		if r.Right != nil {
			r.Value = *findMin(d, cd, ncd, r.Right)
			r.Right, ok = deleteNode(d, r.Value, ncd, r.Right)
		} else if r.Left != nil {
			r.Value = *findMin(d, cd, ncd, r.Left)
			r.Right, ok = deleteNode(d, r.Value, ncd, r.Left)
			r.Left = nil
		} else {
			r = nil
		}
	} else if value.Order(r.Value, cd) == Lesser {
		r.Left, ok = deleteNode(d, value, ncd, r.Left)
	} else {
		r.Right, ok = deleteNode(d, value, ncd, r.Right)
	}
	return r, ok
}

func add[T Comparable[T]](d int, value T, cd int, r *KdNode[T]) bool {
	if value.Dist(r.Value) == 0 {
		return false
	}

	ncd := (cd + 1) % d
	rel := value.Order(r.Value, cd)
	if rel == Lesser {
		if r.Left == nil {
			r.Left = NewKDNode(value)
		} else {
			return add(d, value, ncd, r.Left)
		}
	} else {
		if r.Right == nil {
			r.Right = NewKDNode(value)
		} else {
			return add(d, value, ncd, r.Right)
		}
	}
	return true
}

func nearestNeighbor[T Comparable[T]](d int, v, nn *T, cd int, r *KdNode[T]) *T {
	if r == nil {
		return nil
	}

	var nextBranch, otherBranch *KdNode[T]
	if (*v).Order(r.Value, cd) == Lesser /* [cd] < r.value[cd]*/ {
		nextBranch, otherBranch = r.Left, r.Right
	} else {
		nextBranch, otherBranch = r.Right, r.Left
	}
	ncd := (cd + 1) % d
	nn = nearestNeighbor(d, v, nn, ncd, nextBranch)
	nn = closest(v, nn, &r.Value)

	nearestDist := abs(distance(v, nn))
	dist := abs((*v).DistDim(r.Value, cd))

	if dist <= nearestDist {
		nn = closest(v, nearestNeighbor(d, v, nn, ncd, otherBranch), nn)
	}

	return nn
}

func closest[T Comparable[T]](v, nn1, nn2 *T) *T {
	if nn1 == nil && nn2 == nil {
		panic("Both `nn1` and `nn2` inputs are nil!")
	}

	if nn1 == nil {
		return nn2
	}
	if nn2 == nil {
		return nn1
	}
	if distance(v, nn1) < distance(v, nn2) {
		return nn1
	}
	return nn2
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func distance[T Comparable[T]](src, dst *T) int {
	return (*src).Dist(*dst)
}

func findMin[T Comparable[T]](d, tcd, cd int, r *KdNode[T]) *T {
	if r == nil {
		return nil
	}

	var lMin *T
	var rMin *T
	ncd := (cd + 1) % d
	lMin = findMin(d, tcd, ncd, r.Left)
	if tcd != cd {
		rMin = findMin(d, tcd, ncd, r.Right)
	}
	if lMin == nil && rMin == nil {
		return &r.Value
	} else if lMin == nil {
		if (*rMin).Order(r.Value, tcd) == Lesser {
			return rMin
		}
		return &r.Value
	} else if rMin == nil {
		if (*lMin).Order(r.Value, tcd) == Lesser {
			return lMin
		}
		return &r.Value
	} else {
		// temp := []*T{lMin, rMin, &r.value}
		// sort.Slice(temp, func(i, j int) bool {
		// 	return (*temp[i]).Order(*temp[j], tcd) == Lesser
		// })
		// return temp[0]
		return min(lMin, min(rMin, &r.Value, tcd), tcd)
	}
}

func min[T Comparable[T]](lhs, rhs *T, tcd int) *T {
	if (*lhs).Order(*rhs, tcd) == Lesser {
		return lhs
	}
	return rhs
}

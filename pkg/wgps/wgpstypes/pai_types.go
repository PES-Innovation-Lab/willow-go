package wgpstypes

import (
	"github.com/PES-Innovation-Lab/willow-go/types"
	"github.com/PES-Innovation-Lab/willow-go/utils"
	"golang.org/x/exp/constraints"
)

type FragmentTriple[NamespaceId, SubspaceId constraints.Ordered] struct {
	NamespaceId NamespaceId
	SubspaceId  SubspaceId
	Path        types.Path
}

type FragmentPair[NamespaceId constraints.Ordered] struct {
	NamespaceId NamespaceId
	Path        types.Path
}

type Fragemt interface {
	isFragment()
}

func (FragmentTriple[NamespaceId, SubspaceId]) isFragment() {}
func (FragmentPair[NamespaceId]) isFragment()               {}

type FragmentsComplete[NamespaceId constraints.Ordered] []FragmentPair[NamespaceId]
type FragmentsSelective[NamespaceId, SubspaceId constraints.Ordered] []FragmentTriple[NamespaceId, SubspaceId]

type FragmentSet interface {
	isFragmentSet()
}

func (FragmentsComplete[NamespaceId]) isFragmentSet()               {}
func (FragmentsSelective[NamespeaceId, SubspaceId]) isFragmentSet() {}

type FragmentKitComplete[NamespaceId constraints.Ordered] struct {
	grantedNamespace NamespaceId
	grantedPath      types.Path
}

type FragmentKitSelective[NamespaceId, SubspaceId constraints.Ordered] struct {
	grantedNamespace NamespaceId
	grantedSubspace  SubspaceId
	grantedPath      types.Path
}

type FragmentKit interface {
	isFragmentKit()
}

func (FragmentKitComplete[NamespaceId]) isFragmentKit()              {}
func (FragmentKitSelective[NamespaceId, SubspaceId]) isFragmentKit() {}

type PaiScheme[ReadCapability, PsiScalar any, NamespaceId, SubspaceId constraints.Ordered, K constraints.Unsigned, PsiGroup types.OrderableGeneric] struct {
	fragmentToGroup     func(NamespaceId, SubspaceId) PsiGroup
	getScalar           func() PsiScalar
	scalarMult          func(group PsiGroup, scalar PsiScalar) PsiGroup
	isGroupEqual        func(a PsiGroup, b PsiGroup) bool
	getFragmentKit      func(cap ReadCapability) FragmentKit
	groupMemberEncoding utils.EncodingScheme[PsiGroup, K]
}

type Intersection[PsiGroup any] struct {
	group       PsiGroup
	isComplete  bool
	isSecondary bool
}

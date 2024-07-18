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

type Fragment interface {
	IsFragment()
}

func (FragmentTriple[NamespaceId, SubspaceId]) IsFragment() {}
func (FragmentPair[NamespaceId]) IsFragment()               {}

type FragmentsComplete[NamespaceId constraints.Ordered] []FragmentPair[NamespaceId]
type FragmentsSelective[NamespaceId, SubspaceId constraints.Ordered] []FragmentTriple[NamespaceId, SubspaceId]

type FragmentSet interface {
	IsFragmentSet()
}

func (FragmentsComplete[NamespaceId]) IsFragmentSet()               {}
func (FragmentsSelective[NamespeaceId, SubspaceId]) IsFragmentSet() {}

type FragmentKitComplete[NamespaceId constraints.Ordered] struct {
	GrantedNamespace NamespaceId
	GrantedPath      types.Path
}

type FragmentKitSelective[NamespaceId, SubspaceId constraints.Ordered] struct {
	GrantedNamespace NamespaceId
	GrantedSubspace  SubspaceId
	GrantedPath      types.Path
}

type FragmentKit interface {
	IsFragmentKit()
}

func (FragmentKitComplete[NamespaceId]) IsFragmentKit()              {}
func (FragmentKitSelective[NamespaceId, SubspaceId]) IsFragmentKit() {}

type PaiScheme[ReadCapability, PsiScalar, NamespaceId, SubspaceId constraints.Ordered, K constraints.Unsigned, PsiGroup types.OrderableGeneric] struct {
	FragmentToGroup     func(NamespaceId, SubspaceId) PsiGroup
	GetScalar           func() PsiScalar
	ScalarMult          func(group PsiGroup, scalar PsiScalar) PsiGroup
	IsGroupEqual        func(a PsiGroup, b PsiGroup) bool
	GetFragmentKit      func(cap ReadCapability) FragmentKit
	GroupMemberEncoding utils.EncodingScheme[PsiGroup, K]
}

type Intersection[PsiGroup any] struct {
	Group       PsiGroup
	IsComplete  bool
	IsSecondary bool
}

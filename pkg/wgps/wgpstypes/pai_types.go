package wgpstypes

import (
	"github.com/PES-Innovation-Lab/willow-go/types"
	"github.com/PES-Innovation-Lab/willow-go/utils"
	"golang.org/x/exp/constraints"
)

type FragmentTriple struct {
	NamespaceId types.NamespaceId
	SubspaceId  types.SubspaceId
	Path        types.Path
}

type FragmentPair struct {
	NamespaceId types.NamespaceId
	Path        types.Path
}

type Fragment interface {
	IsFragment()
}

func (FragmentTriple) IsFragment() {}
func (FragmentPair) IsFragment()   {}

type FragmentsComplete []FragmentPair
type FragmentsSelective struct {
	Primary   []FragmentTriple
	Secondary []FragmentPair
}

type FragmentSet interface {
	IsFragmentSet()
}

func (FragmentsComplete) IsFragmentSet()  {}
func (FragmentsSelective) IsFragmentSet() {}

type FragmentKitComplete struct {
	GrantedNamespace types.NamespaceId
	GrantedPath      types.Path
}

type FragmentKitSelective struct {
	GrantedNamespace types.NamespaceId
	GrantedSubspace  types.SubspaceId
	GrantedPath      types.Path
}

type FragmentKit interface {
	IsFragmentKit()
}

func (FragmentKitComplete) IsFragmentKit()  {}
func (FragmentKitSelective) IsFragmentKit() {}

type PaiScheme[ReadCapability, PsiGroup any, PsiScalar int, K constraints.Unsigned] struct {
	FragmentToGroup     func(Fragment) PsiGroup
	GetScalar           func() PsiScalar
	ScalarMult          func(group PsiGroup, scalar PsiScalar) PsiGroup
	IsGroupEqual        func(a PsiGroup, b PsiGroup) bool
	GetFragmentKit      func(cap ReadCapability) FragmentKit
	GroupMemberEncoding utils.EncodingScheme[PsiGroup]
}

type Intersection[PsiGroup any] struct {
	Group       PsiGroup
	IsComplete  bool
	IsSecondary bool
}

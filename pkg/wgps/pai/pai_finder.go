package pai

import (
	"github.com/PES-Innovation-Lab/willow-go/pkg/data_model/datamodeltypes"
	"github.com/PES-Innovation-Lab/willow-go/pkg/wgps/handlestore"
	"github.com/PES-Innovation-Lab/willow-go/pkg/wgps/syncutils"
	"github.com/PES-Innovation-Lab/willow-go/pkg/wgps/wgpstypes"
	"github.com/PES-Innovation-Lab/willow-go/types"
	"github.com/PES-Innovation-Lab/willow-go/utils"
	"golang.org/x/exp/constraints"
)

type PaiFinderOpts[ReadCapability, PsiGroup, SubspaceReadCapability any, PsiScalar int, K constraints.Unsigned] struct {
	NamespaceScheme           datamodeltypes.NamespaceScheme
	PaiScheme                 wgpstypes.PaiScheme[ReadCapability, PsiGroup, PsiScalar, K]
	IntersectionHandlesOurs   handlestore.HandleStore[wgpstypes.Intersection[PsiGroup]]
	IntersectionHandlesTheirs handlestore.HandleStore[wgpstypes.Intersection[PsiGroup]]
}

const (
	BIND_READ_CAP = iota // iota is reset to 0
	REQUEST_SUBSPACE_CAP
	REPLY_READ_CAP
)

// Define an interface that both SubspaceId and any type of ANY_SUBSPACE can satisfy.
type SubspaceOrAny interface{}

const ANY_SUBSPACE = -1

type LocalFragmentInfo[ReadCapability, SubspaceReadCapability any] struct {
	ID             int //set this to 1 if defined, otherwise 0 (default value)
	OnIntersection int
	Authorisation  wgpstypes.ReadAuthorisation[ReadCapability, SubspaceReadCapability]
	Path           types.Path
	Namespace      types.NamespaceId
	Subspace       SubspaceOrAny
}

/** Given `ReadAuthorisation`s, emits the intersected ones  */

type Intersection[ReadCapability, SubspaceReadCapability any] struct {
	NamespaceId       types.NamespaceId
	ReadAuthorisation wgpstypes.ReadAuthorisation[ReadCapability, SubspaceReadCapability]
	Uint64            uint64
}

type BindFragment[PsiGroup any] struct {
	PsiGroup    PsiGroup
	IsSecondary bool
}

type ReplyFragment[PsiGroup any] struct {
	FragmentGrp uint64
	PsiGroup    PsiGroup
}

type SubspaceCapReply[SubspaceReadCapability any] struct {
	Handle                 uint64
	SubspaceReadCapability SubspaceReadCapability
}

type PaiFinder[ReadCapability, PsiGroup, SubspaceReadCapability any, PsiScalar int, K constraints.Unsigned] struct {
	IntersectionHandlesOurs   handlestore.HandleStore[wgpstypes.Intersection[PsiGroup]]
	IntersectionHandlesTheirs handlestore.HandleStore[wgpstypes.Intersection[PsiGroup]]

	IntersectionQueue []Intersection[ReadCapability, SubspaceReadCapability]

	BindFragmentQueue []BindFragment[PsiGroup]

	ReplyFragmentQueue []ReplyFragment[PsiGroup]

	SubspaceCapRequestQueue []uint64

	SubspaceCapReplyQueue []SubspaceCapReply[SubspaceReadCapability]

	FragmentsInfo map[uint64]LocalFragmentInfo[ReadCapability, SubspaceReadCapability]

	NamespaceScheme datamodeltypes.NamespaceScheme

	PaiScheme wgpstypes.PaiScheme[ReadCapability, PsiGroup, PsiScalar, K]

	Scalar PsiScalar

	RequestedSubspaceCapHandles map[uint64]bool
}

func NewPaiFinder[ReadCapability, PsiGroup, SubspaceReadCapability any, PsiScalar int, K constraints.Unsigned](opts PaiFinderOpts[ReadCapability, PsiGroup, SubspaceReadCapability, PsiScalar, K]) *PaiFinder[ReadCapability, PsiGroup, SubspaceReadCapability, PsiScalar, K] {
	return &PaiFinder[ReadCapability, PsiGroup, SubspaceReadCapability, PsiScalar, K]{
		NamespaceScheme:             opts.NamespaceScheme,
		PaiScheme:                   opts.PaiScheme,
		RequestedSubspaceCapHandles: make(map[uint64]bool),
		Scalar:                      opts.PaiScheme.GetScalar(),
		IntersectionHandlesOurs:     opts.IntersectionHandlesOurs,
		IntersectionHandlesTheirs:   opts.IntersectionHandlesTheirs,
		FragmentsInfo:               make(map[uint64]LocalFragmentInfo[ReadCapability, SubspaceReadCapability]),
	}
}

func (p *PaiFinder[ReadCapability, PsiGroup, SubspaceReadCapability, PsiScalar, K]) SubmitAuthorisation(authorisation wgpstypes.ReadAuthorisation[ReadCapability, SubspaceReadCapability]) {
	FragmentKit := p.PaiScheme.GetFragmentKit(authorisation.Capability)
	Fragments := CreateFragmentSet(FragmentKit)

	SubmitFragment := func(fragment wgpstypes.Fragment, isSecondary bool) uint64 {
		unmixed := p.PaiScheme.FragmentToGroup(fragment)
		multiplied := p.PaiScheme.ScalarMult(unmixed, p.Scalar)
		handle := p.IntersectionHandlesOurs.Bind(wgpstypes.Intersection[PsiGroup]{
			Group:       multiplied,
			IsComplete:  false,
			IsSecondary: isSecondary,
		})
		p.BindFragmentQueue = append(p.BindFragmentQueue, BindFragment[PsiGroup]{
			PsiGroup:    multiplied,
			IsSecondary: isSecondary,
		})
		return handle
	}

	if !IsSelectiveFragmentKit(Fragments) {
		for i, fragment := range Fragments.(wgpstypes.FragmentsComplete) {
			namespace, path := fragment.NamespaceId, fragment.Path
			isMostSpecific := i == len(Fragments.(wgpstypes.FragmentsComplete))-1

			groupHandle := SubmitFragment(fragment, false)

			var fragmentInfo LocalFragmentInfo[ReadCapability, SubspaceReadCapability]

			if isMostSpecific {
				// Use Variant1 as an example for the most specific case
				fragmentInfo = LocalFragmentInfo[ReadCapability, SubspaceReadCapability]{
					ID:             1,
					OnIntersection: BIND_READ_CAP,
					Authorisation:  authorisation,
					Path:           path,
					Namespace:      namespace,
					Subspace:       ANY_SUBSPACE,
				}
			} else {
				fragmentInfo = LocalFragmentInfo[ReadCapability, SubspaceReadCapability]{
					ID:             1,
					OnIntersection: REPLY_READ_CAP, // Assuming REPLY_READ_CAP is an int constant
					Authorisation:  authorisation,
					Path:           path,
					Namespace:      namespace,
					Subspace:       ANY_SUBSPACE,
				}
			}

			// Set the fragment info in the map
			p.FragmentsInfo[groupHandle] = fragmentInfo
		}
	} else {
		for i, fragment := range Fragments.(wgpstypes.FragmentsSelective).Primary {
			namespace, subspace, path := fragment.NamespaceId, fragment.SubspaceId, fragment.Path
			isMostSpecific := i == len(Fragments.(wgpstypes.FragmentsSelective).Primary)-1

			groupHandle := SubmitFragment(fragment, false)

			var fragmentInfo LocalFragmentInfo[ReadCapability, SubspaceReadCapability]

			if isMostSpecific {
				// Use Variant1 as an example for the most specific case
				fragmentInfo = LocalFragmentInfo[ReadCapability, SubspaceReadCapability]{
					ID:             1,
					OnIntersection: BIND_READ_CAP,
					Authorisation:  authorisation,
					Path:           path,
					Namespace:      namespace,
					Subspace:       subspace,
				}
			} else {
				// Use Variant2 as an example for the non-specific case
				fragmentInfo = LocalFragmentInfo[ReadCapability, SubspaceReadCapability]{
					ID:             1,
					OnIntersection: REPLY_READ_CAP, // Assuming REPLY_READ_CAP is an int constant
					Authorisation:  authorisation,
					Path:           path,
					Namespace:      namespace,
					Subspace:       subspace,
				}
			}

			// Set the fragment info in the map
			p.FragmentsInfo[groupHandle] = fragmentInfo
		}
	}

	for i, fragment := range Fragments.(wgpstypes.FragmentsSelective).Secondary {
		namespace, path := fragment.NamespaceId, fragment.Path
		isMostSpecific := i == len(Fragments.(wgpstypes.FragmentsSelective).Secondary)-1

		groupHandle := SubmitFragment(fragment, false)

		var fragmentInfo LocalFragmentInfo[ReadCapability, SubspaceReadCapability]

		if isMostSpecific {
			// Use Variant1 as an example for the most specific case
			fragmentInfo = LocalFragmentInfo[ReadCapability, SubspaceReadCapability]{
				ID:             1,
				OnIntersection: BIND_READ_CAP,
				Authorisation:  authorisation,
				Path:           path,
				Namespace:      namespace,
				Subspace:       ANY_SUBSPACE,
			}
		} else {
			// Use Variant2 as an example for the non-specific case
			fragmentInfo = LocalFragmentInfo[ReadCapability, SubspaceReadCapability]{
				ID:             1,
				OnIntersection: REPLY_READ_CAP, // Assuming REPLY_READ_CAP is an int constant
				Authorisation:  authorisation,
				Path:           path,
				Namespace:      namespace,
				Subspace:       ANY_SUBSPACE,
			}
		}

		// Set the fragment info in the map
		p.FragmentsInfo[groupHandle] = fragmentInfo
	}
}

func (p *PaiFinder[ReadCapability, PsiGroup, SubspaceReadCapability, PsiScalar, K]) ReceivedBind(groupMember PsiGroup, isSecondary bool) {
	multiplied := p.PaiScheme.ScalarMult(groupMember, p.PaiScheme.GetScalar())
	handle := p.IntersectionHandlesOurs.Bind(wgpstypes.Intersection[PsiGroup]{
		Group:       multiplied,
		IsComplete:  true,
		IsSecondary: isSecondary,
	})
	p.ReplyFragmentQueue = append(p.ReplyFragmentQueue, ReplyFragment[PsiGroup]{
		FragmentGrp: handle,
		PsiGroup:    multiplied,
	})
	p.CheckForIntersections(handle, false)
}

func (p *PaiFinder[ReadCapability, PsiGroup, PsiScalar, SubspaceReadCapability, K]) ReceivedReply(handle uint64, groupMember PsiGroup) {
	_, found := p.IntersectionHandlesOurs.Get(handle)
	if !found {
		//throw an error
	}
	p.IntersectionHandlesOurs.Update(handle, wgpstypes.Intersection[PsiGroup]{
		Group:       groupMember,
		IsComplete:  true,
		IsSecondary: false,
	})
	p.CheckForIntersections(handle, true)
}

func (p *PaiFinder[ReadCapability, PsiGroup, SubspaceReadCapability, PsiScalar, K]) ReceivedSubspaceCapRequest(handle uint64, isReply bool) {
	result, found := p.IntersectionHandlesTheirs.Get(handle)
	if !found {
		//throw an error
	}
	for ourHandle, intersection := range p.IntersectionHandlesOurs {
		//Need to tell AC to incorporate the iterator into handle store
		if !intersection.IsComplete {
			continue
		}
		if !p.PaiScheme.IsGroupEqual(intersection.Group, result.Group) {
			continue
		}
		var FragmentInfo LocalFragmentInfo[ReadCapability, SubspaceReadCapability] = p.FragmentsInfo[ourHandle]

		if FragmentInfo.ID != 1 {
			//throw an error
		}

		if !syncutils.IsSubspaceReadAuthorisation(FragmentInfo.Authorisation) {
			continue
		}
		p.SubspaceCapReplyQueue = append(p.SubspaceCapReplyQueue, SubspaceCapReply[SubspaceReadCapability]{
			Handle:                 handle,
			SubspaceReadCapability: FragmentInfo.Authorisation.SubspaceCapability, //need to see how to do this
		})
	}
}

func (p *PaiFinder[ReadCapability, PsiGroup, SubspaceReadCapability, PsiScalar, K]) ReceivedVerifiedSubspaceCapReply(handle uint64, namespace types.NamespaceId) {
	if !p.RequestedSubspaceCapHandles[handle] {
		//throw a willow error
	}
	delete(p.RequestedSubspaceCapHandles, handle)
	_, choice := p.IntersectionHandlesTheirs.Get(handle)
	if !choice {
		//throw an error
	}
	fragmentInfo := p.FragmentsInfo[handle]
	if fragmentInfo.ID != 1 {
		//throw an error
	}
	if !p.NamespaceScheme.IsEqual(fragmentInfo.Namespace, namespace) { //need to see how to do this
		//throw an error
	}
	p.IntersectionQueue = append(p.IntersectionQueue, Intersection[ReadCapability, SubspaceReadCapability]{
		NamespaceId:       namespace,
		ReadAuthorisation: fragmentInfo.Authorisation,
		Uint64:            handle,
	})
}

func (p *PaiFinder[ReadCapability, PsiGroup, SubspaceReadCapability, PsiScalar, K]) CheckForIntersections(handle uint64, ours bool) {
	var storeToGetHandleFrom handlestore.HandleStore[wgpstypes.Intersection[PsiGroup]]
	if ours {
		storeToGetHandleFrom = p.IntersectionHandlesOurs
	} else {
		storeToGetHandleFrom = p.IntersectionHandlesTheirs
	}
	var storeToCheckAgainst handlestore.HandleStore[wgpstypes.Intersection[PsiGroup]]
	if ours {
		storeToCheckAgainst = p.IntersectionHandlesTheirs
	} else {
		storeToCheckAgainst = p.IntersectionHandlesOurs
	}

	intersection, choice := storeToGetHandleFrom.Get(handle)
	if !choice {
		//throw an error
	}
	if !intersection.IsComplete {
		return
	}
	for otherHandle, otherIntersection := range storeToCheckAgainst {
		if !otherIntersection.IsComplete {
			continue
		}
		if intersection.IsSecondary && otherIntersection.IsSecondary {
			continue
		}
		if !p.PaiScheme.IsGroupEqual(intersection.Group, otherIntersection.Group) {
			continue
		}
		var ourHandle uint64
		if ours {
			ourHandle = handle
		} else {
			ourHandle = otherHandle
		}

		fragmentInfo := p.FragmentsInfo[ourHandle]

		if fragmentInfo.ID != 1 {
			//throw an error
		}
		if fragmentInfo.OnIntersection == BIND_READ_CAP {
			p.IntersectionQueue = append(p.IntersectionQueue, Intersection[ReadCapability, SubspaceReadCapability]{
				NamespaceId:       fragmentInfo.Namespace,
				ReadAuthorisation: fragmentInfo.Authorisation,
				Uint64:            ourHandle,
			})
		} else if fragmentInfo.OnIntersection == REQUEST_SUBSPACE_CAP {
			p.RequestedSubspaceCapHandles[ourHandle] = true
			p.SubspaceCapRequestQueue = append(p.SubspaceCapRequestQueue, handle)
		}
	}
}

func (p *PaiFinder[ReadCapability, PsiGroup, PsiScalar, SubspaceReadCapability, K]) GetHandleOuterArea(handle uint64) types.Area {
	fragmentInfo := p.FragmentsInfo[handle]
	if fragmentInfo.ID != 1 {
		//throw an error
	}
	return types.Area{
		Subspace_id: fragmentInfo.Subspace.(types.SubspaceId),
		Path:        fragmentInfo.Path,
		Times:       types.Range[uint64]{Start: 0, OpenEnd: true},
	}
}

func (p *PaiFinder[ReadCapability, PsiGroup, SubspaceReadCapability, PsiScalar, K]) ReceivedReadCapForIntersection(theirIntersectionHandle uint64) {
	theirIntersection, choice := p.IntersectionHandlesTheirs.Get(theirIntersectionHandle)
	if !choice {
		//throw an error
	}
	for ourHandle, ourIntersection := range p.IntersectionHandlesOurs {
		if !ourIntersection.IsComplete {
			continue
		}
		if ourIntersection.IsSecondary && theirIntersection.IsSecondary {
			continue
		}
		if !p.PaiScheme.IsGroupEqual(ourIntersection.Group, theirIntersection.Group) {
			continue
		}
		fragmentInfo := p.FragmentsInfo[ourHandle]
		if fragmentInfo.ID != 1 {
			//throw an error
		}
		if fragmentInfo.OnIntersection == REPLY_READ_CAP {
			p.IntersectionQueue = append(p.IntersectionQueue, Intersection[ReadCapability, SubspaceReadCapability]{
				NamespaceId:       fragmentInfo.Namespace,
				ReadAuthorisation: fragmentInfo.Authorisation,
				Uint64:            ourHandle,
			})
		}
	}
}

func (p *PaiFinder[ReadCapability, PsiGroup, SubspaceReadCapability, PsiScalar, K]) GetIntersectionPrivy(handle uint64, ours bool) wgpstypes.ReadCapPrivy {
	var storeToGetHandleFrom handlestore.HandleStore[wgpstypes.Intersection[PsiGroup]]
	if ours {
		storeToGetHandleFrom = p.IntersectionHandlesOurs
	} else {
		storeToGetHandleFrom = p.IntersectionHandlesTheirs
	}
	var storeToCheckAgainst handlestore.HandleStore[wgpstypes.Intersection[PsiGroup]]
	if ours {
		storeToCheckAgainst = p.IntersectionHandlesTheirs
	} else {
		storeToCheckAgainst = p.IntersectionHandlesOurs
	}
	intersection, choice := storeToGetHandleFrom.Get(handle)
	if !choice {
		//throw an error
	}
	// Here we are looping through the whole contents of the handle store because...
	// otherwise we need to build a special handle store just for intersections.
	// Which we might do one day, but I'm not convinced it's worth it yet.
	for otherHandle, otherIntersection := range storeToCheckAgainst {
		if !otherIntersection.IsComplete {
			continue
		}
		if intersection.IsSecondary && otherIntersection.IsSecondary {
			continue
		}
		if !p.PaiScheme.IsGroupEqual(intersection.Group, otherIntersection.Group) {
			continue
		}
		// If there is an intersection, check what we have to do!
		var ourHandle uint64
		if ours {
			ourHandle = handle
		} else {
			ourHandle = otherHandle
		}
		fragmentInfo := p.FragmentsInfo[ourHandle]
		if fragmentInfo.ID != 1 {
			//throw an error
		}
		outer := p.GetHandleOuterArea(ourHandle)
		return wgpstypes.ReadCapPrivy{
			Namespace: fragmentInfo.Namespace,
			Outer:     outer,
		}
	}
	//throw an error
	return wgpstypes.ReadCapPrivy{} // This is a placeholder
}

func (p *PaiFinder[ReadCapability, PsiGroup, SubspaceReadCapability, PsiScalar, K]) FragmentBinds() <-chan BindFragment[PsiGroup] {
	// Assuming p.BindFragmentQueue is a channel or can be iterated over in some way
	// This channel is where we'll send our FragmentBind structs
	ch := make(chan BindFragment[PsiGroup])

	go func() {
		defer close(ch) // Ensure the channel is closed when we're done sending data

		// Simulate iterating over bindFragmentQueue similar to the JavaScript for-await loop
		// This is a placeholder loop. You'll need to adapt this to your actual data source.
		for bind := range p.BindFragmentQueue { // Assuming BindFragmentQueue is a channel
			Group := p.BindFragmentQueue[bind].PsiGroup
			IsSecondary := p.BindFragmentQueue[bind].IsSecondary
			// Send a FragmentBind struct to the channel
			ch <- BindFragment[PsiGroup]{
				PsiGroup:    Group,
				IsSecondary: IsSecondary,
			}
		}
	}()

	return ch // Return the channel to the caller
}

func (p *PaiFinder[ReadCapability, PsiGroup, SubspaceReadCapability, PsiScalar, K]) FragmentReplies() <-chan ReplyFragment[PsiGroup] {
	ch := make(chan ReplyFragment[PsiGroup])

	go func() {
		defer close(ch)

		for bind := range p.ReplyFragmentQueue {
			Group := p.ReplyFragmentQueue[bind].PsiGroup
			FragmentGrp := p.ReplyFragmentQueue[bind].FragmentGrp

			ch <- ReplyFragment[PsiGroup]{
				PsiGroup:    Group,
				FragmentGrp: FragmentGrp,
			}
		}
	}()

	return ch
}

func (p *PaiFinder[ReadCapability, PsiGroup, SubspaceReadCapability, PsiScalar, K]) Intersections() <-chan Intersection[ReadCapability, SubspaceReadCapability] {
	ch := make(chan Intersection[ReadCapability, SubspaceReadCapability])

	go func() {
		defer close(ch)

		for bind := range p.IntersectionQueue {
			NamespaceId := p.IntersectionQueue[bind].NamespaceId
			ReadAuthorisation := p.IntersectionQueue[bind].ReadAuthorisation
			Uint64 := p.IntersectionQueue[bind].Uint64

			ch <- Intersection[ReadCapability, SubspaceReadCapability]{
				NamespaceId:       NamespaceId,
				ReadAuthorisation: ReadAuthorisation,
				Uint64:            Uint64,
			}
		}
	}()

	return ch
}

func (p *PaiFinder[ReadCapability, PsiGroup, SubspaceReadCapability, PsiScalar, K]) SubspaceCapRequests() <-chan uint64 {
	ch := make(chan uint64)

	go func() {
		defer close(ch)

		for bind := range p.SubspaceCapRequestQueue {
			handle := p.SubspaceCapRequestQueue[bind]

			ch <- handle
		}
	}()

	return ch
}

func (p *PaiFinder[ReadCapability, PsiGroup, SubspaceReadCapability, PsiScalar, K]) SubspaceCaReplies() <-chan SubspaceCapReply[SubspaceReadCapability] {
	ch := make(chan SubspaceCapReply[SubspaceReadCapability])

	go func() {
		defer close(ch)

		for bind := range p.SubspaceCapReplyQueue {
			handle := p.SubspaceCapReplyQueue[bind].Handle
			subspaceReadCapability := p.SubspaceCapReplyQueue[bind].SubspaceReadCapability

			ch <- SubspaceCapReply[SubspaceReadCapability]{
				Handle:                 handle,
				SubspaceReadCapability: subspaceReadCapability,
			}
		}
	}()

	return ch
}

func CreateFragmentSet(kit wgpstypes.FragmentKit) wgpstypes.FragmentSet {
	primaryFragment := []wgpstypes.FragmentTriple{}
	secondaryFragment := []wgpstypes.FragmentPair{}
	switch v := kit.(type) {
	case wgpstypes.FragmentKitSelective:
		prefixes := utils.PrefixesOf(kit.(wgpstypes.FragmentKitSelective).GrantedPath)
		for _, prefix := range prefixes {
			primaryFragmentInfo := wgpstypes.FragmentTriple{
				NamespaceId: v.GrantedNamespace,
				SubspaceId:  v.GrantedSubspace,
				Path:        prefix,
			}
			secondaryFragmentInfo := wgpstypes.FragmentPair{
				NamespaceId: v.GrantedNamespace,
				Path:        prefix,
			}
			primaryFragment = append(primaryFragment, primaryFragmentInfo)
			secondaryFragment = append(secondaryFragment, secondaryFragmentInfo)
		}
		return wgpstypes.FragmentsSelective{
			Primary:   primaryFragment,
			Secondary: secondaryFragment,
		}
	}
	prefixes := utils.PrefixesOf(kit.(wgpstypes.FragmentKitComplete).GrantedPath)
	pairs := []wgpstypes.FragmentPair{}
	for _, prefix := range prefixes {
		pairsInfo := wgpstypes.FragmentPair{
			NamespaceId: kit.(wgpstypes.FragmentKitComplete).GrantedNamespace,
			Path:        prefix,
		}
		pairs = append(pairs, pairsInfo)
	}
	return wgpstypes.FragmentsSelective{
		Secondary: pairs,
	}
}

func IsSelectiveFragmentKit(set wgpstypes.FragmentSet) bool {
	_, ok := set.(wgpstypes.FragmentsSelective)
	return ok
}

func IsFragmentTriple(fragment wgpstypes.Fragment) bool {
	_, ok := fragment.(*wgpstypes.FragmentTriple)
	return ok
}

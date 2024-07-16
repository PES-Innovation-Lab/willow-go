package pai

import (
	"github.com/PES-Innovation-Lab/willow-go/pkg/data_model/datamodeltypes"
	"github.com/PES-Innovation-Lab/willow-go/pkg/wgps"
	"github.com/PES-Innovation-Lab/willow-go/pkg/wgps/wgpstypes"
	"github.com/PES-Innovation-Lab/willow-go/types"
	"golang.org/x/exp/constraints"
)

type PaiFinderOpts[ReadCapability, PsiGroup, PsiScalar, NamespaceId, SubspaceId constraints.Ordered, K constraints.Unsigned] struct {
	NamespaceScheme           datamodeltypes.NamespaceScheme[NamespaceId, K]
	PaiScheme                 wgpstypes.PaiScheme[ReadCapability, PsiGroup, PsiScalar, NamespaceId, SubspaceId, K]
	IntersectionHandlesOurs   wgps.HandleStore[wgpstypes.Intersection[PsiGroup]]
	IntersectionHandlesTheirs wgps.HandleStore[wgpstypes.Intersection[PsiGroup]]
}

const (
	BIND_READ_CAP = iota // iota is reset to 0
	REQUEST_SUBSPACE_CAP
	REPLY_READ_CAP
)

// Define the interface that all variants will implicitly implement
type LocalFragmentInfo interface{}

type Variant1[ReadCapability, SubspaceReadCapability, NamespaceId, SubspaceId constraints.Ordered] struct {
	OnIntersection int
	Authorisation  wgpstypes.ReadAuthorisation[ReadCapability, SubspaceReadCapability]
	Path           types.Path
	Namespace      NamespaceId
	Subspace       SubspaceId
}

type Variant2[ReadCapability, SubspaceReadCapability, NamespaceId, SubspaceId constraints.Ordered] struct {
	OnIntersection int
	Authorisation  wgpstypes.ReadAuthorisation[ReadCapability, SubspaceReadCapability]
	Path           types.Path
	Namespace      NamespaceId
	Subspace       SubspaceId
}

type Variant3[ReadCapability, SubspaceReadCapability, NamespaceId, SubspaceId constraints.Ordered] struct {
	OnIntersection int
	Authorisation  wgpstypes.ReadAuthorisation[ReadCapability, SubspaceReadCapability]
	Path           types.Path
	Namespace      NamespaceId
	Subspace       SubspaceId
}

/** Given `ReadAuthorisation`s, emits the intersected ones  */
type Intersection[ReadCapability, SubspaceReadCapability, NamespaceId constraints.Ordered] struct {
	NamespaceId       NamespaceId
	ReadAuthorisation wgpstypes.ReadAuthorisation[ReadCapability, SubspaceReadCapability]
	Uint64            uint64
}
type BindFragment[PsiGroup constraints.Ordered] struct {
	PsiGroup    PsiGroup
	IsSecondary bool
}
type ReplyFragment[PsiGroup constraints.Ordered] struct {
	FragmentGrp uint64
	PsiGroup    PsiGroup
}
type SubspaceCapReply[SubspaceReadCapability constraints.Ordered] struct {
	Handle                 uint64
	SubspaceReadCapability SubspaceReadCapability
}

type PaiFinder[ReadCapability, PsiGroup, PsiScalar, SubspaceReadCapability, NamespaceId, SubspaceId constraints.Ordered, K constraints.Unsigned] struct {
	IntersectionHandlesOurs   wgps.HandleStore[wgpstypes.Intersection[PsiGroup]]
	IntersectionHandlesTheirs wgps.HandleStore[wgpstypes.Intersection[PsiGroup]]

	IntersectionQueue []Intersection[ReadCapability, SubspaceReadCapability, NamespaceId]

	BindFragmentQueue []BindFragment[PsiGroup]

	ReplyFragmentQueue []ReplyFragment[PsiGroup]

	SubspaceCapRequestQueue []uint64

	SubspaceCapReplyQueue []SubspaceCapReply[SubspaceReadCapability]

	FragmentsInfo map[uint64]LocalFragmentInfo

	NamespaceScheme datamodeltypes.NamespaceScheme[NamespaceId, K]

	PaiScheme wgpstypes.PaiScheme[ReadCapability, PsiGroup, PsiScalar, NamespaceId, SubspaceId, K]

	Scalar PsiScalar

	RequestedSubspaceCapHandles map[uint64]bool
}

func NewPaiFinder[ReadCapability, PsiGroup, PsiScalar, SubspaceReadCapability, NamespaceId, SubspaceId constraints.Ordered, K constraints.Unsigned](opts PaiFinderOpts[ReadCapability, PsiGroup, PsiScalar, NamespaceId, SubspaceId, K]) *PaiFinder[ReadCapability, PsiGroup, PsiScalar, SubspaceReadCapability, NamespaceId, SubspaceId, K] {
	return &PaiFinder[ReadCapability, PsiGroup, PsiScalar, SubspaceReadCapability, NamespaceId, SubspaceId, K]{
		NamespaceScheme:             opts.NamespaceScheme,
		PaiScheme:                   opts.PaiScheme,
		RequestedSubspaceCapHandles: make(map[uint64]bool),
		Scalar:                      opts.PaiScheme.GetScalar(),
		IntersectionHandlesOurs:     opts.IntersectionHandlesOurs,
		IntersectionHandlesTheirs:   opts.IntersectionHandlesTheirs,
		FragmentsInfo:               make(map[uint64]LocalFragmentInfo),
	}
}

func (p *PaiFinder[ReadCapability, PsiGroup, PsiScalar, SubspaceReadCapability, NamespaceId, SubspaceId, K]) SubmitAuthorisation(authorisation wgpstypes.ReadAuthorisation[ReadCapability, SubspaceReadCapability]) {
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

	if !isSelectiveFragmentKit(Fragments) {
		for i, fragment := range Fragments {
			namespace, path := fragment[0], fragment[1]
			isMostSpecific := i == len(Fragments)-1

			groupHandle := SubmitFragment(fragment, false)

			var fragmentInfo LocalFragmentInfo

			if isMostSpecific {
				// Use Variant1 as an example for the most specific case
				fragmentInfo = Variant1[ReadCapability, SubspaceReadCapability, NamespaceId, SubspaceId]{
					OnIntersection: BIND_READ_CAP,
					Authorisation:  authorisation,
					Path:           path,
					Namespace:      namespace,
					Subspace:       ANY_SUBSPACE,
				}
			} else {
				// Use Variant2 as an example for the non-specific case
				fragmentInfo = Variant2[ReadCapability, SubspaceReadCapability, NamespaceId, SubspaceId]{
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
		for i, fragment := range Fragments {
			namespace, subspace, path := fragment[0], fragment[1], fragment[2]
			isMostSpecific := i == len(Fragments)-1

			groupHandle := SubmitFragment(fragment, false)

			var fragmentInfo LocalFragmentInfo

			if isMostSpecific {
				// Use Variant1 as an example for the most specific case
				fragmentInfo = Variant1[ReadCapability, SubspaceReadCapability, NamespaceId, SubspaceId]{
					OnIntersection: BIND_READ_CAP,
					Authorisation:  authorisation,
					Path:           path,
					Namespace:      namespace,
					Subspace:       subspace,
				}
			} else {
				// Use Variant2 as an example for the non-specific case
				fragmentInfo = Variant2[ReadCapability, SubspaceReadCapability, NamespaceId, SubspaceId]{
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

	for i, fragment := range Fragments {
		namespace, path := fragment[0], fragment[1]
		isMostSpecific := i == len(Fragments)-1

		groupHandle := SubmitFragment(fragment, false)

		var fragmentInfo LocalFragmentInfo

		if isMostSpecific {
			// Use Variant1 as an example for the most specific case
			fragmentInfo = Variant1[ReadCapability, SubspaceReadCapability, NamespaceId, SubspaceId]{
				OnIntersection: BIND_READ_CAP,
				Authorisation:  authorisation,
				Path:           path,
				Namespace:      namespace,
				Subspace:       ANY_SUBSPACE,
			}
		} else {
			// Use Variant2 as an example for the non-specific case
			fragmentInfo = Variant2[ReadCapability, SubspaceReadCapability, NamespaceId, SubspaceId]{
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

func (p *PaiFinder[ReadCapability, PsiGroup, PsiScalar, SubspaceReadCapability, NamespaceId, SubspaceId, K]) ReceivedBind(groupMember PsiGroup, isSecondary bool) {
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

func (p *PaiFinder[ReadCapability, PsiGroup, PsiScalar, SubspaceReadCapability, NamespaceId, SubspaceId, K]) ReceivedReply(handle uint64, groupMember PsiGroup) {
	intersection, _ := p.IntersectionHandlesOurs.Get(handle)
	if !intersection {
		//check how this can be done
	}
	p.IntersectionHandlesOurs.Update(handle, wgpstypes.Intersection[PsiGroup]{
		Group:       groupMember,
		IsComplete:  true,
		IsSecondary: false,
	})
	p.CheckForIntersections(handle, true)
}

func (p *PaiFinder[ReadCapability, PsiGroup, PsiScalar, SubspaceReadCapability, NamespaceId, SubspaceId, K]) ReceivedSubspaceCapRequest(handle uint64, isReply bool) {
	result, _ := p.IntersectionHandlesTheirs.Get(handle)
	if !result {
		//check how this can be done
	}
	for ourHandle, intersection := range p.IntersectionHandlesOurs {
		//Need to tell AC to incorporate the iterator into handle store
		if !intersection.IsComplete {
			continue
		}
		if !p.PaiScheme.IsGroupEqual(intersection.Group, result.Group) {
			continue
		}
		FragmentInfo := p.FragmentsInfo[ourHandle]

		if !FragmentInfo {
			//check how this can be done
		}
		p.SubspaceCapReplyQueue = append(p.SubspaceCapReplyQueue, SubspaceCapReply[SubspaceReadCapability]{
			Handle:                 handle,
			SubspaceReadCapability: FragmentInfo.Authorisation.SubspaceCapability, //need to see how to do this
		})
	}
}

func (p *PaiFinder[ReadCapability, PsiGroup, PsiScalar, SubspaceReadCapability, NamespaceId, SubspaceId, K]) ReceivedVerifiedSubspaceCapReply(handle uint64, namespace NamespaceId) {
	if !p.RequestedSubspaceCapHandles[handle] {
		//throw a willow error
	}
	delete(p.RequestedSubspaceCapHandles, handle)
	result, _ := p.IntersectionHandlesTheirs.Get(handle)
	if !result {
		//see how this can be done
	}
	fragmentInfo := p.FragmentsInfo[handle]
	if !fragmentInfo {
		//see how this can be done
	}
	if !p.NamespaceScheme.IsEqual(fragmentInfo.Namespace, namespace) { //need to see how to do this
		//throw an error
	}
	p.IntersectionQueue = append(p.IntersectionQueue, Intersection[ReadCapability, SubspaceReadCapability, NamespaceId]{
		NamespaceId:       namespace,
		ReadAuthorisation: fragmentInfo.Authorisation,
		Uint64:            handle,
	})
}

func (p *PaiFinder[ReadCapability, PsiGroup, PsiScalar, SubspaceReadCapability, NamespaceId, SubspaceId, K]) CheckForIntersections(handle uint64, ours bool) {
	var storeToGetHandleFrom wgps.HandleStore[wgpstypes.Intersection[PsiGroup]]
	if ours {
		storeToGetHandleFrom = p.IntersectionHandlesOurs
	} else {
		storeToGetHandleFrom = p.IntersectionHandlesTheirs
	}
	var storeToCheckAgainst wgps.HandleStore[wgpstypes.Intersection[PsiGroup]]
	if ours {
		storeToCheckAgainst = p.IntersectionHandlesTheirs
	} else {
		storeToCheckAgainst = p.IntersectionHandlesOurs
	}

	intersection, _ := storeToGetHandleFrom.Get(handle)
	if !intersection {
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

		if !fragmentInfo {
			//throw an error
		}
		if fragmentInfo.OnIntersection == BIND_READ_CAP {
			p.IntersectionQueue = append(p.IntersectionQueue, Intersection[ReadCapability, SubspaceReadCapability, NamespaceId]{
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

func (p *PaiFinder[ReadCapability, PsiGroup, PsiScalar, SubspaceReadCapability, NamespaceId, SubspaceId, K]) GetHandleOuterArea(handle uint64) types.Area[SubspaceId] {
	fragmentInfo := p.FragmentsInfo[handle]
	if !fragmentInfo {
		//throw an error
	}
	return types.Area[SubspaceId]{
		Subspace_id: fragmentInfo.Subspace,
		Path:        fragmentInfo.Path,
		Times:       types.Range[uint64]{Start: 0, OpenEnd: true},
	}
}

func (p *PaiFinder[ReadCapability, PsiGroup, PsiScalar, SubspaceReadCapability, NamespaceId, SubspaceId, K]) ReceivedReadCapForIntersection(theirIntersectionHandle uint64) {
	theirIntersection, _ := p.IntersectionHandlesTheirs.Get(theirIntersectionHandle)
	if !theirIntersection {
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
		if !fragmentInfo {
			//throw an error
		}
		if fragmentInfo.OnIntersection == REPLY_READ_CAP {
			p.IntersectionQueue = append(p.IntersectionQueue, Intersection[ReadCapability, SubspaceReadCapability, NamespaceId]{
				NamespaceId:       fragmentInfo.Namespace,
				ReadAuthorisation: fragmentInfo.Authorisation,
				Uint64:            ourHandle,
			})
		}
	}
}

func (p *PaiFinder[ReadCapability, PsiGroup, PsiScalar, SubspaceReadCapability, NamespaceId, SubspaceId, K]) GetIntersectionPrivy(handle uint64, ours bool) wgpstypes.ReadCapPrivy[NamespaceId, SubspaceId] {
	var storeToGetHandleFrom wgps.HandleStore[wgpstypes.Intersection[PsiGroup]]
	if ours {
		storeToGetHandleFrom = p.IntersectionHandlesOurs
	} else {
		storeToGetHandleFrom = p.IntersectionHandlesTheirs
	}
	var storeToCheckAgainst wgps.HandleStore[wgpstypes.Intersection[PsiGroup]]
	if ours {
		storeToCheckAgainst = p.IntersectionHandlesTheirs
	} else {
		storeToCheckAgainst = p.IntersectionHandlesOurs
	}
	intersection, _ := storeToGetHandleFrom.Get(handle)
	if !intersection {
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
		if !fragmentInfo {
			//throw an error
		}
		outer := p.GetHandleOuterArea(ourHandle)
		return wgpstypes.ReadCapPrivy[NamespaceId, SubspaceId]{
			Namespace: fragmentInfo.Namespace,
			Outer:     outer,
		}
	}
	//throw an error
	return wgpstypes.ReadCapPrivy[NamespaceId, SubspaceId]{} // This is a placeholder
}

//NEED TO DO ALL THE ASYNC ITERABLES

// Assuming prefixesOf is a function that takes a path and returns a slice of types.Path
// func prefixesOf(path types.Path) []types.Path {
//     // Implementation goes here
// }

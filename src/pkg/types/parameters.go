package types

import "golang.org/x/exp/constraints"

// TODO: Depends on encoding tyopes

type PathParams[T constraints.Unsigned] struct {
	/*
	  Setting up Path parameters, we are setting each type to be unisgned as these parameters cannot be negative
	  we are also not setting it to signed and not a fixed uint32 so that if the user does not want the params to be that long
	  we can save some space by using uint8 if required.
	*/
	MaxComponentcount  T
	MaxComponentLength T
	MaxPathLength      T
}

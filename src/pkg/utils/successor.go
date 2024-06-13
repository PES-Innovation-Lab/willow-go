package utils

func successorPath(path Path, scheme PathScheme) Path {
	if len(path) == 0 {
		nextPath := Path{make([]byte, 1)}

		if isValidPath(nextPath, scheme) {
			return nextPath
		}

		return nil
	}

	workingPath := make(Path, len(path))
	copy(workingPath, path)

	for i := len(path) - 1; i >= 0; i-- {
		component := workingPath[i]

		simplestNextComponent := append(component, 0)

		simplestNextPath := make(Path, len(path))
		copy(simplestNextPath, path)
		simplestNextPath[i] = simplestNextComponent

		if isValidPath(simplestNextPath, scheme) {
			return simplestNextPath
		}

		incrementedComponent := successorBytesFixedWidth(component)

		if incrementedComponent != nil {
			nextPath := append(path[:i], incrementedComponent)
			return nextPath
		}

		workingPath = workingPath[:len(workingPath)-1]
	}

	if len(workingPath) == 0 {
		return nil
	}

	return workingPath
}

func successorPrefix(path Path) Path {
	if len(path) == 0 {
		return nil
	}

	workingPath := make(Path, len(path))
	copy(workingPath, path)

	for i := len(path) - 1; i >= 0; i-- {
		component := workingPath[i]

		incrementedComponent := successorBytesFixedWidth(component)

		if incrementedComponent != nil {
			nextPath := append(path[:i], incrementedComponent)
			return nextPath
		}

		if len(component) == 0 {
			nextPath := append(path[:i], []byte{0})
			return nextPath
		}

		workingPath = workingPath[:len(workingPath)-1]
	}

	if len(workingPath) == 0 {
		return nil
	}

	return workingPath
}

func successorBytesFixedWidth(bytes []byte) []byte {
	newBytes := make([]byte, len(bytes))
	copy(newBytes, bytes)

	didIncrement := false

	for i := len(newBytes) - 1; i >= 0; i-- {
		byteVal := newBytes[i]

		if byteVal >= 255 {
			newBytes[i] = 0
			continue
		}

		if !didIncrement {
			newBytes[i] = byteVal + 1
			didIncrement = true
			break
		}
	}

	if !didIncrement {
		return nil
	}

	return newBytes
}

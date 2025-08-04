package bind

const (
	query uint8 = iota
	path
	body
)

type binding struct {
	kind         uint8
	qualifier    string
	optional     bool
	defaultValue string
}

func Query(queryParameter string) binding {
	return binding{kind: query, qualifier: queryParameter}
}

func QueryOpt(queryParameter string, defaultValue string) binding {
	return binding{kind: query, qualifier: queryParameter, optional: true, defaultValue: defaultValue}
}

func Path(pathParameter string) binding {
	return binding{kind: path, qualifier: pathParameter}
}

func PathOpt(pathParameter string, defaultValue string) binding {
	return binding{kind: path, qualifier: pathParameter, optional: true, defaultValue: defaultValue}
}

func Body(bodyType string) binding {
	return binding{kind: path, qualifier: bodyType}
}

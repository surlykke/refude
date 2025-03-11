package bind

type binding struct {
	query string
	req   bool
	def   string
	path  string
	body  string
}

func Q(queryParam string) binding {
	return binding{query: queryParam}
}

func QR(queryParam string) binding {
	return binding{query: queryParam, req: true}
}

func QD(queryParam string, def string) binding {
	return binding{query: queryParam, def: def}
}

func P(pathParam string) binding {
	return binding{path: pathParam}
}

func Json() binding {
	return binding{body: "json"}
}

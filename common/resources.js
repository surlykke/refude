// Copyright (c) 2015, 2016, 2017 Christian Surlykke
//
// This file is part of the refude project. 
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
import {combinedUrl, combinedUrls, iconServiceUrl, doHttp} from '../common/utils'

let dressup = (url, resource) => {
	resource.url =  url
	resource.IconUrl = resource.IconUrl ? combinedUrl(resource.url, resource.IconUrl) :
	 			       resource.IconName ? iconServiceUrl(resource.IconName) :
                       undefined
}

let MakeCollection = (service, pathPrefix, onUpdate) => {
	let indexUrl = "http://localhost:7938/" + service + pathPrefix + "/"
	let notifyUrl = "http://localhost:7938/" + service + "/notify"

	let index = []
	let resourceMap = new Map()

	let resourceCollection = []

	let update = () => {
		resourceCollection.length = 0
		index.forEach(url => {
			if (resourceMap[url]) {
				resourceCollection.push(resourceMap[url])
			}
		})
		resourceCollection.sort((res1, res2) => (res2.RelevanceHint || 0) - (res1.RelevanceHint || 0))
		onUpdate()
	}

	let setResource = (url, res) => {
		if (url === indexUrl) {
			index = []
			res.forEach(path => {index.push(combinedUrl(indexUrl, path))})
			index.forEach(url => {
				if (!resourceMap[url]) {
					fetchResource(url)
				}
			})
		}
		else {
			dressup(url, res)
			resourceMap[url] = res
		}
		update()
	}

	let fetchResource = (url) => {
		doHttp(url).then(res => { setResource(url, res) })
		           .catch(err => { console.log("Error fetching", url, err)})

	}

	let fetchIfRelevant = url => {
		if (url === indexUrl || index.includes(url)) { fetchResource(url) }
	}

	let evtSource;

	let connect = () => {
		console.log("Source connecting to ", notifyUrl)
		evtSource = new EventSource(notifyUrl)
		evtSource.onopen = () => {
			fetchResource(indexUrl)
		}

		evtSource.onerror = event => {
			resourceMap = new Map()
			if (evtSource.readyState === 2) {
				setTimeout(() => {connect()}, 5000)
			}
			update()
		}

		evtSource.addEventListener("resource-updated", event =>  {
			fetchIfRelevant(combinedUrl(notifyUrl, event.data))
		})
	}

	connect()
	return resourceCollection
}

let MakeResource = (service, path, onUpdate) => {
	let resourceUrl = "http://localhost:7938/" + service + path
	let notifyUrl = "http://localhost:7938/" + service + "/notify"
	let resource = {}

	let setResource = (res) => {
		for (let k in resource) {
			resource[k] = undefined
		}
		for (let k in res) {
			resource[k] = res[k]
		}
		dressup(resourceUrl, resource)
		onUpdate()
	}

	let fetch = () => {
		doHttp(resourceUrl)
			.then(setResource)
            .catch(err => {
				console.log("Error fetching", url, err)
				setResource({})
			})

	}

	let evtSource;

	let connect = () => {
		console.log("Source connecting to ", notifyUrl)
		evtSource = new EventSource(notifyUrl)
		evtSource.onopen = () => {
			fetch()
		}

		evtSource.onerror = event => {
			if (evtSource.readyState === 2) {
				setTimeout(() => {connect()}, 5000)
			}
			setResource({})
		}

		evtSource.addEventListener("resource-updated", event =>  {
			if (resourceUrl === combinedUrl(notifyUrl, event.data)) fetch()
		})

		evtSource.addEventListener("resource-added", event =>  {
			if (resourceUrl === combinedUrl(notifyUrl, event.data)) fetch()
		})

		evtSource.addEventListener("resource-removed", event =>  {
			if (resourceUrl === combinedUrl(notifyUrl, event.data)) setResource({})
		})
	}

	connect()
	return resource
}

export {MakeResource, MakeCollection}

import {combinedUrl, combinedUrls, iconServiceUrl, doHttp} from '../common/utils'

let dressup = (url, resource) => {
	resource.url =  url
	resource.IconUrl = resource.IconUrl ? combinedUrl(resource.url, resource.IconUrl) :
	 			       resource.IconName ? iconServiceUrl(resource.IconName) :
                       undefined
}

export function MakeCollection(service, pathPrefix, onUpdate) {
	let indexUrl = "http://localhost:7938/" + service + pathPrefix + "/"
	let notifyUrl = "http://localhost:7938/" + service + "/notify"
	let searchTerm = ""

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
		doHttp(url).then( res => { setResource(url, res) })
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

let http = require('http')

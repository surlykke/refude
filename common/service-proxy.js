import {combinedUrl, combinedUrls, iconServiceUrl, doHttp} from '../common/utils'

export function MakeServiceProxy(service, directory) {
	let indexUrl = service + directory
	let notifyUrl = service + "/notify"
	
	let subscribers = []

	let proxy = {
		get: url => { return resourceMap[url] },
		index: () => { return resourceMap[indexUrl] || [] },
		resources: () => { return proxy.index().map(url => resourceMap[url]).filter(res => res)},
		subscribe: subscriber => {subscribers.push(subscriber)},
		unsubscribe: subscriber => {subscribers = subscribers.filter(s => s !== subscriber)},
		indexUrl: indexUrl,
	}

	let publish = url => {
		subscribers.forEach(subscriber => {subscriber(url)})
	}

	let dressup = (url, resource) => {
		resource.url =  url
		resource.IconUrl = resource.IconUrl ? combinedUrl(resource.url, resource.IconUrl) :
		 			       resource.IconName ? iconServiceUrl(resource.IconName) :
                           undefined
	}

	let setResource = (url, res) => {
		if (url === indexUrl) {
			res = combinedUrls(url, res)
			res.filter(url => !resourceMap[url]).forEach(url => {fetchResource(url)})
		}
		else {
			dressup(url, res)
		}

		resourceMap[url] = res
		publish(url)
	}

	let fetchResource = (url) => {
		doHttp(url).then( res => { setResource(url, res) })
	}

	let fetchIfRelevant = url => {
		if (url === indexUrl || proxy.index().includes(url)) {
			fetchResource(url)
		}
	}

	let resourceMap = new Map()
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
		}

		evtSource.addEventListener("resource-updated", event =>  {
			fetchIfRelevant(combinedUrl(notifyUrl, event.data))
		})

		evtSource.addEventListener("resource-added", event => {
			fetchIfRelevant(combinedUrl(notifyUrl, event.data))
		})

		evtSource.addEventListener("resource-removed", event => {
			let url = combinedUrl(notifyUrl, event.data)
			if (resourceMap[url]) {
				delete resourceMap[url]
				publish(url)
			}
		})
	}

	connect()
	return proxy
}

let http = require('http')

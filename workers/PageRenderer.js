// Import the Cloudflare Workers KV and HTMLRewriter libraries
const { Queue } = require('cloudflare-workers-kv')
const { HTMLRewriter } = require('html-rewriter')

// Define your Cloudflare Worker
addEventListener('fetch', event => {
  event.respondWith(handleRequest(event))
})

async function handleRequest(event) {
  // Get a reference to the "RenderedPages" queue
  const renderedPages = new Queue('RenderedPages')

  // Wait for the next message in the "FetchedLinks" queue
  const fetchedLinks = new Queue('FetchedLinks')
  await fetchedLinks.awaitQueue()

  // Process the next message in the "FetchedLinks" queue
  while (true) {
    const message = await fetchedLinks.get()
    if (!message) {
      break
    }

    // Parse the link from the message
    const row = JSON.parse(message.value)
    const link = row.link

    // Fetch the web page for the link
    const response = await fetch(link)

    // Handle dynamic web pages using HTMLRewriter
    const content = await response.text()
    const contentType = response.headers.get('content-type')
    if (contentType.includes('text/html')) {
      const rewriter = new HTMLRewriter()
        .on('a', new AttributeRewriter('href'))
        .on('img', new AttributeRewriter('src'))
      const transformed = rewriter.transform(content)
      await renderedPages.put(transformed)
    } else {
      await renderedPages.put(content)
    }
  }

  return new Response('Processed all messages', { status: 200 })
}

class AttributeRewriter {
  constructor(attributeName) {
    this.attributeName = attributeName
  }

  element(element) {
    const attribute = element.getAttribute(this.attributeName)
    if (attribute) {
      element.setAttribute(this.attributeName, this.rewriteUrl(attribute))
    }
  }

  rewriteUrl(url) {
    // Add your custom URL rewriting logic here
    return url
  }
}


/*

Write a Cloudflare worker that does the following:

- Gets triggered when there is a new message in a queue named "FetchedLinks"
- Extracts a field called Link in the message
- Check if this md5 hash for this link is present in the VisitedPages KV namespace. If yes, exit. if no, make a new key for that link
- Renders the webpage content using that link
- Does not call an external API for this task
- Once rendered, computes a md5 hash
- Check if this hash is present in the VisitedPages KV namespace
- If present, no action will be taken after that
- If not present, add it to the KV namespace
- Save the webpage link and the content as JSON in a queue called RenderedPages

*/
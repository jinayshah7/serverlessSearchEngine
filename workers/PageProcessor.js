// Import the Cloudflare Workers KV, HTMLRewriter, and PlanetScale libraries
const { Queue } = require('cloudflare-workers-kv')
const planetScale = require('planetscale')

// Set up the PlanetScale connection
const client = planetScale.createClient({
  apiKey: '<your PlanetScale API key>',
  organization: '<your PlanetScale organization>',
  project: '<your PlanetScale project>',
  database: '<your PlanetScale database>',
  table: '<your PlanetScale table>'
})

// Define your Cloudflare Worker
addEventListener('fetch', event => {
  event.respondWith(handleRequest(event))
})

async function handleRequest(event) {
  // Get a reference to the "RenderedPages" and "ProcessedPages" queues
  const renderedPages = new Queue('RenderedPages')
  const processedPages = new Queue('ProcessedPages')

  // Wait for the next message in the "RenderedPages" queue
  await renderedPages.awaitQueue()

  // Process the next message in the "RenderedPages" queue
  while (true) {
    const message = await renderedPages.get()
    if (!message) {
      break
    }

    // Parse the HTML content from the message
    const content = message.value

    // Extract all links from the HTML content
    const links = []
    const re = /href="([^"]*)"/g
    let match
    while ((match = re.exec(content)) !== null) {
      const link = match[1]
      if (link.startsWith('http')) {
        links.push(link)
      }
    }

    // Save the links to the "ProcessedPages" queue
    for (const link of links) {
      await processedPages.put(link)

      // Insert a new row into the PlanetScale table for the link
      const row = {
        source: message.key,
        destination: link
      }
      await client.table('<your PlanetScale table>').insert(row)
    }
  }

  return new Response('Processed all messages', { status: 200 })
}


/*

Write a Cloudflare worker that does the following:

- Gets triggered when there is a new message in a queue named "RenderedPages"
- Assume the queue is already created, just import it
- Extracts a field called link and content from the message. we'll call that link sourceLink
- Extract all links from the content
- For each link found, make a new row in the table called LinkGraph in PlanetScale
- Two columns: source and destination. Each row will look like this (sourceLink, link found in the content)
- Construct one single SQL query and make only one API call
- For each link found, check if it is present in the VisitedPages KV namespace
- If yes, continue. If not, make a new key for the link.
- Put the sourcelink and content in a queue called "ProcessedPages"

*/
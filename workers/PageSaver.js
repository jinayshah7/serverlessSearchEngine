// Import the Cloudflare Workers KV and Algolia libraries
const { Queue } = require('cloudflare-workers-kv')
const algoliasearch = require('algoliasearch')

// Set up the Algolia connection
const client = algoliasearch('<your Algolia application ID>', '<your Algolia API key>')
const index = client.initIndex('<your Algolia index>')

// Define your Cloudflare Worker
addEventListener('fetch', event => {
  event.respondWith(handleRequest(event))
})

async function handleRequest(event) {
  // Get a reference to the "ProcessedPages" queue
  const processedPages = new Queue('ProcessedPages')

  // Wait for the next message in the "ProcessedPages" queue
  await processedPages.awaitQueue()

  // Process the next message in the "ProcessedPages" queue
  while (true) {
    const message = await processedPages.get()
    if (!message) {
      break
    }

    // Send the message to Algolia for storage and indexing
    await index.addObject({
      url: message.value
    })
  }

  return new Response('Processed all messages', { status: 200 })
}

/*

Write a Cloudflare worker that does the following:

- Gets triggered when there is a new message in a queue named "ProcessedPages"
- Extracts a field called Link and Content in the message
- Send this data to Algolia for indexing, use Link as the ID

*/
import { Queue } from 'https://cdn.jsdelivr.net/npm/@workers-queue/queue'

addEventListener('fetch', event => {
  event.respondWith(handleRequest(event.request))
})

async function handleRequest(request) {
  // Authenticate with PlanetScale API
  const apiKey = 'YOUR_PLANETSCALE_API_KEY'
  const headers = { 'Authorization': `Bearer ${apiKey}` }

  // Get random row from UnvisitedLinks
  const unvisitedUrl = 'https://api.planetscale.com/v1alpha1/projects/YOUR_PROJECT/databases/YOUR_DATABASE/tables/UnvisitedLinks/query'
  const unvisitedQuery = 'SELECT * FROM UnvisitedLinks ORDER BY RAND() LIMIT 1'
  const unvisitedResponse = await fetch(unvisitedUrl, { method: 'POST', headers, body: JSON.stringify({ query: unvisitedQuery }) })
  const unvisitedJson = await unvisitedResponse.json()
  const unvisitedLink = unvisitedJson.rows[0].link

  // Transfer link to VisitedLinks
  const visitedUrl = 'https://api.planetscale.com/v1alpha1/projects/YOUR_PROJECT/databases/YOUR_DATABASE/tables/VisitedLinks/query'
  const visitedQuery = `INSERT INTO VisitedLinks (link) VALUES ("${unvisitedLink}")`
  const visitedResponse = await fetch(visitedUrl, { method: 'POST', headers, body: JSON.stringify({ query: visitedQuery }) })

  // Save result to Cloudflare queue
  const queueName = 'FetchedLinks'
  const queue = new Queue(queueName, {
    accountId: 'YOUR_ACCOUNT',
    durable: true,
  })
  await queue.add({ link: unvisitedLink })

  return new Response('Success', { status: 200 })
}

/*

Write a Cloudflare worker that does the following:

- Call the PlanetScale API to get a random row from the table called UnvisitedLinks
- It will call the API to transfer that link to another table called VisitedLinks
- Do both tasks using a single SQL query and API call.
- Save it to a Cloudflare queue with the name FetchedLinks


*/
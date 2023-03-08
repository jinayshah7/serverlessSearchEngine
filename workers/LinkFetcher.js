import { Queue } from 'https://cdn.jsdelivr.net/npm/@workers-queue/queue'

addEventListener('scheduled', event => {
  event.respondWith(handleScheduled(event.request))
})

const cronSchedule = '0 * * * *'
const timezone = 'UTC'

const scheduleOptions = {cron: cronSchedule, timezone: timezone}
const scheduledEvent = new ScheduledEvent('hourly-cron', scheduleOptions)
scheduledEvent.schedule()

async function handleScheduled(request) {
  // Authenticate with PlanetScale API
  const apiKey = await SECRETS.PLANETSCALE_API_KEY
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
  const queue = new Queue(queueName)
  await queue.add({ link: unvisitedLink })

  return new Response('Success', { status: 200 })
}

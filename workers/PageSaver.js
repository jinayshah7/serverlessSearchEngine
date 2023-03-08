// Import the Cloudflare Workers KV and Algolia libraries
import { Queue } from 'cloudflare-workers-kv';
import algoliasearch from 'algoliasearch';

addEventListener('scheduled', event => {
  event.respondWith(handleScheduled())
})

const cronSchedule = '0 * * * *'
const timezone = 'UTC'

const scheduleOptions = {cron: cronSchedule, timezone: timezone}
const scheduledEvent = new ScheduledEvent('hourly-cron', scheduleOptions)
scheduledEvent.schedule()


const ALGOLIA_APPLICATION_ID = await SECRETS.ALGOLIA_APPLICATION_ID;
const ALGOLIA_APPLICATION_KEY = await SECRETS.ALGOLIA_APPLICATION_KEY;
const ALGOLIA_INDEX = await SECRETS.ALGOLIA_INDEX;
const INPUT_QUEUE_NAME = 'ProcessedPages';

const client = algoliasearch(ALGOLIA_APPLICATION_ID, ALGOLIA_APPLICATION_KEY);
const index = client.initIndex(ALGOLIA_INDEX);

async function handleScheduled() {
  // Get a reference to the "ProcessedPages" queue
  const page = getPageFromQueue()

    await index.addObject({
      url: page.link,
      content: page.content
    })

  return new Response('Page Saved', { status: 200 })
}

async function getPageFromQueue() {
  const queue = new Queue(INPUT_QUEUE_NAME);
  const page = await queue.pop();
  const { link, content } = JSON.parse(page);
  return { link, content };
}

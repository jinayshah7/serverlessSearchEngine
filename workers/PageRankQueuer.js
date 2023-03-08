
import { createClient } from '@planetscale/cli';
import { Queue } from 'cloudflare-workers';

addEventListener('scheduled', event => {
  event.respondWith(handleScheduled())
})

const cronSchedule = '0 * * * *'
const timezone = 'UTC'

const scheduleOptions = {cron: cronSchedule, timezone: timezone}
const scheduledEvent = new ScheduledEvent('hourly-cron', scheduleOptions)
scheduledEvent.schedule()

// Define the name of the PageRankState KV namespace
const PAGE_RANK_STATE_KV_NAMESPACE = 'PageRankState';
const PLANETSCALE_API_TOKEN = await SECRETS.PLANETSCALE_API_KEY;
const DB_NAME = await SECRETS.DB_NAME;
const OUTPUT_QUEUE_NAME = 'PageRankQueue';

const client = createClient({
  token: PLANETSCALE_API_TOKEN,
});

const db = client.database(DB_NAME);
const kv = new KVNamespace(PAGE_RANK_STATE_KV_NAMESPACE);

addEventListener('fetch', event => {
  event.respondWith(handleRequest(event.request));
});

async function handleRequest(request) {

  const iterationNumber = getIterationNumber();
  link = await getLinkFromDatabase(iterationNumber)
  await saveMessageToQueue(link)

  return new Response("Link queued for page rank processing", { status: 200 })
}

async function saveMessageToQueue(link){
  const queue = new Queue(OUTPUT_QUEUE_NAME);
  await queue.push(link);
}

async function getIterationNumber(){
  parseInt(await kv.get('Iteration Number'));
}

async function getLinkFromDatabase(iterationNumber){

  const sqlQuery = `SELECT * FROM VisitedLinks WHERE iteration_number <= ${iterationNumber} ORDER BY RAND() LIMIT 1`;

  try {

    const visitedTable = await db.getTable(VISITED_LINKS_TABLE_NAME);
    const result = await visitedTable.query(sqlQuery);
    if (result.length == 0) {
        await kv.set('Iteration Number', iterationNumber+1);
        return {};
    }
    const link = result[0];

    await visitedTable.update({
      where: { id: link.id },
      data: { iteration_number: link.iterationNumber + 1 }
    });

    return link

  } catch (error) {
    return {};
  }
}
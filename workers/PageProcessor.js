import Queue from 'queue';
import { generateMD5Hash } from 'crypto';
import { createClient } from '@planetscale/cli';
import { KVNamespace } from '@cloudflare/workers-kv'

addEventListener('scheduled', event => {
  event.respondWith(handleScheduled())
})

const cronSchedule = '0 * * * *'
const timezone = 'UTC'

const scheduleOptions = {cron: cronSchedule, timezone: timezone}
const scheduledEvent = new ScheduledEvent('hourly-cron', scheduleOptions)
scheduledEvent.schedule()

const PLANETSCALE_API_TOKEN = await SECRETS.PLANETSCALE_API_KEY;
const DB_NAME = await SECRETS.DB_NAME;
const LINKGRAPH_TABLE_NAME = 'LinkGraph';
const UNVISITED_LINKS_TABLE_NAME = 'LinkGraph';
const INPUT_QUEUE_NAME = 'RenderedPages';
const OUTPUT_QUEUE_NAME = 'ProcessedPages';
const VISITED_PAGE_HASHES_KV_NAMESPACE = 'VisitedPagesHashes';


const client = createClient({
  token: PLANETSCALE_API_TOKEN,
});

const db = client.database(DB_NAME);
const kv = new KVNamespace(VISITED_PAGE_HASHES_KV_NAMESPACE);

async function handleScheduled() {

  let message = await getContentFromQueue()
  const links = extractLinks(message.content);
  await addLinksToLinkGraph(message.link, links)
  await saveLinksToDatabase(links)
  await saveMessageToQueue(message)

  return new Response('Done');
}

async function saveMessageToQueue(message){
  const processedQueue = new Queue(OUTPUT_QUEUE_NAME);
  await processedQueue.push(message);
}

async function saveLinksToDatabase(links){
  const unvisitedTable = await db.getTable(UNVISITED_LINKS_TABLE_NAME);
  await Promise.all(unvisitedLinks.map(async (link) => {
    await unvisitedTable.insert({ link });
  }));
}

function extractLinks(content) {
  const linkRegex = /<a href="(.*?)"/gi;
  const matches = content.matchAll(linkRegex);
  const links = Array.from(matches, (match) => match[1]);
  return links;
}

async function getContentFromQueue() {
  const queue = new Queue(INPUT_QUEUE_NAME);
  const message = await queue.pop();
  const { link, content } = JSON.parse(message);
  return { link, content };
}

async function addLinksToLinkGraph(sourceLink, links){
  const newRows = links.map((link) => {
    const row = {
      source: sourceLink,
      destination: link,
    };
    return row;
  });
  
  await db.getTable(LINKGRAPH_TABLE_NAME).insert(newRows);

  const unvisitedLinks = [];
  for (const link of links) {
    const linkHash = generateMD5Hash(link);
    if (!kv[linkHash]) {
      unvisitedLinks.push(link);
      kv[linkHash] = true;
    }
  }
}
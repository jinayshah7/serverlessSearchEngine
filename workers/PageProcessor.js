import Queue from 'queue';
import { generateMD5Hash } from 'crypto';
import { createClient } from '@planetscale/cli';

addEventListener('fetch', (event) => {
  event.respondWith(handleRequest(event.request));
});

const PLANETSCALE_API_TOKEN = '<your PlanetScale API token>';
const PLANETSCALE_DB_NAME = '<your database name>';
const PLANETSCALE_TABLE_NAME = 'LinkGraph';

const KV_NAMESPACE_NAME = '<your KV namespace name>';

const client = createClient({
  token: PLANETSCALE_API_TOKEN,
});

const db = client.database(PLANETSCALE_DB_NAME);

const kv = await MY_KV_NAMESPACE.get('VisitedPages');

async function handleRequest(request) {
  const queue = new Queue('RenderedPages');
  
  const message = await queue.pop();
  const { link: sourceLink, content } = JSON.parse(message);

  const links = extractLinks(content);
  
  const newRows = links.map((link) => {
    const row = {
      source: sourceLink,
      destination: link,
    };
    return row;
  });
  
  await db.getTable(PLANETSCALE_TABLE_NAME).insert(newRows);

  const unvisitedLinks = [];
  for (const link of links) {
    const linkHash = generateMD5Hash(link);
    if (!kv[linkHash]) {
      unvisitedLinks.push(link);
      kv[linkHash] = true;
    }
  }

  const processedMessage = JSON.stringify({ link: sourceLink, content });
  const processedQueue = new Queue('ProcessedPages');
  await processedQueue.push(processedMessage);

  const unvisitedTable = await db.getTable('UnvisitedLinks');
  await Promise.all(unvisitedLinks.map(async (link) => {
    await unvisitedTable.insert({ link });
  }));

  return new Response('Done');
}

function extractLinks(content) {
  const linkRegex = /<a href="(.*?)"/gi;
  const matches = content.matchAll(linkRegex);
  const links = Array.from(matches, (match) => match[1]);
  return links;
}

async function getKVNamespace() {
  const namespace = await MY_KV.getNamespace(KV_NAMESPACE_NAME);
  return namespace;
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
- For each link found, check if it's md5 hash is present in the VisitedPages KV namespace.
- If link is present, continue to the next step. If link is not found, get the md5 hash of that link and make a new key in KV.
- Add that link the to the UnvisitedLinks table in PlanetScale
- Put the sourcelink and content in a queue called "ProcessedPages"

*/
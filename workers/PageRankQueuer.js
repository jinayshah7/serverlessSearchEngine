/*

Write a Cloudflare worker that does the following:

- Check the PageRankState KV namespace for a key called "Iteration Number", let's call it N. Get N first.
- Call the PlanetScale API to get a random row from the table called VisitedLinks. I want one row with column iteration_number <= N
- Along with that, also set the iteration_number column for that row to N+1.
- Do both tasks using a single SQL query and API call.
- Save the row contents to a Cloudflare queue with the name PageRankQueue


*/

// Import the required modules
import { MY_PLANETSCALE_API_KEY, MY_PLANETSCALE_DB } from './secrets.js';
import { Client } from '@planetscale/client';
import { Queue } from 'cloudflare-workers';

// Initialize the PlanetScale API client
const planetScaleApiClient = new Client({ api_key: MY_PLANETSCALE_API_KEY }).table(MY_PLANETSCALE_DB.database, 'VisitedLinks');

// Define the name of the PageRankState KV namespace
const pageRankStateNamespace = 'PageRankState';

addEventListener('fetch', event => {
  event.respondWith(handleRequest(event.request));
});

async function handleRequest(request) {
  // Get the current iteration number from the PageRankState KV namespace
  const iterationNumber = parseInt(await PAGE_RANK_STATE.get('Iteration Number'));

  // Construct the SQL query to get a random row from the VisitedLinks table with iteration_number <= N
  const sqlQuery = `SELECT * FROM VisitedLinks WHERE iteration_number <= ${iterationNumber} ORDER BY RAND() LIMIT 1`;

  try {
    // Call the PlanetScale API to execute the SQL query and get a random row from the VisitedLinks table
    const result = await planetScaleApiClient.query(sqlQuery);

    // Get the row data from the API response
    const rowData = result[0];

    // Set the iteration_number column for the selected row to N+1
    await planetScaleApiClient.update({
      where: { id: rowData.id },
      data: { iteration_number: iterationNumber + 1 }
    });

    // Get the PageRankQueue queue using the Cloudflare SDK
    const queue = new Queue(MY_CLOUDFLARE_ACCOUNT, 'PageRankQueue');

    // Save the row contents to the PageRankQueue queue
    await queue.put(JSON.stringify(rowData));

    return new Response(`Random row from VisitedLinks table saved to PageRankQueue: ${JSON.stringify(rowData)}`, { status: 200 });
  } catch (error) {
    return new Response(`Error: ${error.message}`, { status: 500 });
  }
}
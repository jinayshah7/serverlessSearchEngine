/*

Write a Cloudflare worker that does the following:

- Check the PageRankState KV namespace for a key called "Iteration Number", let's call it N. Get N first.
- Call the PlanetScale API to get a random row from the table called VisitedLinks. I want one row with column iteration_number <= N
- Along with that, also set the iteration_number column for that row to N+1.
- Do both tasks using a single SQL query and API call.
- Save the row contents to a Cloudflare queue with the name PageRankQueue


*/
package main

// If there are new messages in Kafka, get the first one.
// Find ID, get the document using RPC
// Calculate pageRank for this document
// Call the rpc which will keep track of the state and total residual score
// Find out IDs of all documents linked from this document
// Keep the ones that are not processed yet
// add all of them to the Kafka topic

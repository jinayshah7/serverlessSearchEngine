import * as pulumi from "@pulumi/pulumi";
import * as cloudflare from "@pulumi/cloudflare";
import * as fs from "fs";

const domain = "jinay.io";

// Create Cloudflare Workers
const workerNames = ["LinkFetcher", "PageRenderer", "PageProcessor", "PageSaver", "PageRankQueuer", "PageRankProcessor"];
const workers = workerNames.map(name => {
    const scriptFile = fs.readFileSync(`./workers/${name}.js`, "utf8");
    return new cloudflare.Worker(name, {
        script: new pulumi.asset.StringAsset(scriptFile),
        routes: [`https://${domain}/${name}`],
    });
});

// Create Cloudflare Queues
const queueNames = ["FetchedLinks", "RenderedPages", "ProcessedPages", "PageRankeQueue"];
const queues = queueNames.map(name => new cloudflare.Queue(name, {
    maxSize: 1000,
    retentionSeconds: 7200,
}));

// Create KV Namespaces
const kvNamespaceNamess = ["VisitedPageHashes", "VisitedLinkHashes", "PageRankState"];
const kvNamespaces = kvNamespaceNamess.map(name => new cloudflare.KVNamespace(name, {}));

// Create Cloudflare Pages site
const pagesSite = new cloudflare.PagesSite("pages", {
    zoneId: cloudflare.getZone({ name: domain }).then(zone => zone.id),
    custom404Path: "/404.html",
    engine: "v8",
    buildCommand: "npm install && npm run build",
    previewCommand: "npm run start",
    environmentVariables: {
        NODE_ENV: "production",
    },
    routes: [{
        pattern: "/",
        path: "build/index.html",
    }],
});

// Output some useful information
export const workerUrls = workers.map(worker => worker.defaultRouteUrl);
export const queueIds = queues.map(queue => queue.id);
export const kvNamespaceNames = kvNamespaces.map(kvNamespace => kvNamespace.name);
export const pagesSiteUrl = pulumi.interpolate`https://${domain}`;

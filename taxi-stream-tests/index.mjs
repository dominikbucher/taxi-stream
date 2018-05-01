// Start with "node --experimental-modules index.mjs" or "npm run app". Requires Node v9+.

/**
 * This script tests the throughput of the taxi and client WebSocket streams. It simply records
 * how many messages each stream receives in a one second window, and reports the number to the
 * console. No magic :).
 */
import WebSocket from "ws";

// const wsClients = new WebSocket('ws://129.132.127.249:8080/ws-clients');
// const wsTaxis = new WebSocket('ws://129.132.127.249:8080/ws');
const wsClients = new WebSocket('ws://127.0.0.1:8080/ws-clients');
const wsTaxis = new WebSocket('ws://127.0.0.1:8080/ws');

let countClients = 0;
let countTaxis = 0;
let lastTimeClients = process.hrtime();
let lastTimeTaxis = process.hrtime();

wsClients.on('message', function incoming(data) {
    countClients++;
    const diff = process.hrtime(lastTimeClients);
    if (diff[0] >= 1) {
        console.log("Clients:", countClients);
        countClients = 0;
        lastTimeClients = process.hrtime();
    }
});

wsTaxis.on('message', function incoming(data) {
    countTaxis++;
    const diff = process.hrtime(lastTimeTaxis);
    if (diff[0] >= 1) {
        console.log("Taxis:", countTaxis);
        countTaxis = 0;
        lastTimeTaxis = process.hrtime();
    }
});
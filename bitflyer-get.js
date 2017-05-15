#!/usr/bin/env node
const request = require('request');
const crypto = require('crypto');

const key = process.env['BITFLYER_API_KEY'];
const secret = process.env['BITFLYER_API_SECRET'];

const timestamp = Date.now().toString();
const method = 'GET';
const path = (process.argv[2] || "/v1/me/getbalance");
const text = timestamp + method + path;
const sign = crypto.createHmac('sha256', secret).update(text).digest('hex');

const options = {
    url: 'https://api.bitflyer.jp' + path,
    method: method,
    headers: {
        'ACCESS-KEY': key,
        'ACCESS-TIMESTAMP': timestamp,
        'ACCESS-SIGN': sign
    }
};
request(options, function (err, response, payload) {
    console.log(payload);
});

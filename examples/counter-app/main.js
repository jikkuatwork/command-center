// Counter App - Serverless JavaScript Example
// Demonstrates the KV store for persistent data

let count = db.get("count") || 0;
count++;
db.set("count", count);

res.send(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Counter App</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            min-height: 100vh;
            display: flex;
            justify-content: center;
            align-items: center;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
        }
        .container {
            text-align: center;
            background: white;
            padding: 60px 80px;
            border-radius: 20px;
            box-shadow: 0 20px 60px rgba(0,0,0,0.3);
        }
        h1 { color: #333; margin-bottom: 10px; }
        .count {
            font-size: 120px;
            font-weight: bold;
            color: #667eea;
            line-height: 1;
            margin: 20px 0;
        }
        p { color: #666; }
        .refresh {
            margin-top: 30px;
            padding: 15px 40px;
            font-size: 18px;
            background: #667eea;
            color: white;
            border: none;
            border-radius: 10px;
            cursor: pointer;
            text-decoration: none;
            display: inline-block;
        }
        .refresh:hover { background: #5a6fd6; }
    </style>
</head>
<body>
    <div class="container">
        <h1>Visitor Counter</h1>
        <div class="count">${count}</div>
        <p>Total page views</p>
        <a href="/" class="refresh">Refresh</a>
    </div>
</body>
</html>`);

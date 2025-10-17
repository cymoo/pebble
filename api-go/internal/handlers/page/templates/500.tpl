{{define "500"}}
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>500 - Internal Server Error</title>
  <style>
    body {
      font-family: system-ui, -apple-system, sans-serif;
      display: flex;
      justify-content: center;
      align-items: center;
      min-height: 100vh;
      margin: 0;
      background: #f5f5f5;
    }
    .container {
      text-align: center;
      padding: 2rem;
    }
    h1 {
      font-size: 6rem;
      margin: 0;
      color: #d32f2f;
    }
    p {
      font-size: 1.5rem;
      color: #666;
    }
    a {
      color: #0066cc;
      text-decoration: none;
    }
    a:hover {
      text-decoration: underline;
    }
  </style>
</head>
<body>
  <div class="container">
    <h1>500</h1>
    <p>Internal Server Error</p>
    <a href="/">Go Home</a>
  </div>
</body>
</html>
{{end}}
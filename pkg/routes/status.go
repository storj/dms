package routes

var StatusPageTemplate = `<!doctype html>
<html lang="en">
  <head>
    <!-- Required meta tags -->
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">

    <!-- Bootstrap CSS -->
    <link rel="stylesheet" href="https://stackpath.bootstrapcdn.com/bootstrap/4.3.1/css/bootstrap.min.css" integrity="sha384-ggOyR0iXCbMQv3Xipma34MD+dH/1fQ784/j6cY/iJTQUOhcWr7x9JvoRxT2MZw1T" crossorigin="anonymous">

    <title>DMS: Status</title>
  </head>
  <body>
	<div class="container-fluid">
		<table class="table">
			<thead>
				<tr>
					<th scope="col">env</th>
					<th scope="col">last_updated</th>
				<tr>
			<thead/>

			<tbody>
				{{ range $key, $val := . }}
				<tr>
					<td>{{ $key }}</td>
					<td>{{ $val }}</td>
				</tr>
				{{ end }}
	</div>
  </body>
</html>`

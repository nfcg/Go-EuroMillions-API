# Go-EuroMillions-API

EuroMillions-API is a lightweight HTTP server that serves EuroMillions drawing results from an SQLite database.
The server can be configured with various options and supports multiple endpoints and output formats.

### Building and Running

To build the executable, use the following command:

```bash
go build go-euromillions-api.go
````

To run the server, use:

```bash
./go-euromillions-api [options]
```

The server starts on port `8080` by default.

### Command-Line Options

| Flag | Shorthand | Description | Default Value |
| :--- | :--- | :--- | :--- |
| `--database` | `-d` | Path to the SQLite database file. | `./euromillions.db`|
| `--verbose` | | Enable verbose logging for requests. | `false`|
| `--log-file` | `-l` | Path to a log file. Output is to the console by default. | (empty)|
| `--version` | `-v` | Show the application version. | `false`|
| `--help` | `-h` | Show the application help message. | `false`|

### API Endpoints

The API supports the `?format` URL query parameter to specify the output format, with valid options being `json` (default), `xml`, and `plaintext`.

  * **GET `/`**: Returns the latest drawing result.
  * **GET `/results`**: Returns all drawing results from the database.
  * **GET `/results/latest`**: Returns the latest drawing result. Example: `/results/latest?format=json`.
  * **GET `/results/date/{date}`**: Searches for a result on a specific date. The date format is `YYYY-MM-DD`. Example: `/results/date/2024-01-15`.
  * **GET `/results/year/{year}`**: Returns all results for a specific year. The year format is `YYYY`. Example: `/results/year/2023`.
  * **GET `/results/month/{month}`**: Returns all results for a specific month and year. The month format is `YYYY-MM`. Example: `/results/month/2024-03`.


## Database Updater (`go-euromillions-api-update.go`)
A command-line tool to scrape the latest EuroMillions results from various sources and update a local SQLite database.

### Building and Running

To build the executable, use:

```bash
go build go-euromillions-api-update.go
```

To run the updater, use:

```bash
./go-euromillions-api-update [options]
```

### Command-Line Options
**Flags**:

  * `--database, -d`: Path to the SQLite database file. (required)
  * `--site, -s`: The site ID to update (`1`, `2`, `4`) or `'all'` to run all. (required)
  * `--verbose, -v`: Enable verbose logging. (default: `false`)
  * `--output, -o`: Path to a log file. Output is to the console by default.

### Supported Sites

  * `1`: https://www.euromilhoes.com/
  * `2`: https://www.euro-millions.com/results
  * `4`: https://www.euromillones.com/result-euromillions.asp
  * `all`: Runs all.
  
### Usage Examples

  * **Update from a single site (ID 1):**

    ```bash
    ./go-euromillions-api-update --database /path/to/euromillions.db --site 1
    ```

  * **Update from all supported sites:**

    ```bash
    ./go-euromillions-api-update --database /path/to/euromillions.db --site all --verbose
    ```

  * **Update with output to a log file:**

    ```bash
    ./go-euromillions-api-update -d /path/to/euromillions.db -s all -o update.log
    ```


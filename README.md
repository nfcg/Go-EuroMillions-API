# Go-EuroMillions-API

EuroMillions API is a lightweight HTTP server written in Go, it serves EuroMillions drawing results from an SQLite database.  
The server can be configured with various options and supports multiple endpoints and output formats. 

----

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

<hr> 

### Command-Line Options

| Flag | Shorthand | Description | Default Value |
| :--- | :--- | :--- | :--- |
| `--database` | `-d` | Path to the SQLite database file. | `./euromillions.db`|
| `--verbose` | | Enable verbose logging for requests. | `false`|
| `--log-file` | `-l` | Path to a log file. Output is to the console by default. | (empty)|
| `--version` | `-v` | Show the application version. | `false`|
| `--help` | `-h` | Show the application help message. | `false`|

<hr> 

### API Endpoints

The API supports the `?format` URL query parameter to specify the output format, with valid options being `json` (default), `xml`, and `plaintext`.

  * **GET `/`**: Returns the latest drawing result.
  * **GET `/results`**: Returns all drawing results from the database.
  * **GET `/results/latest`**: Returns the latest drawing result. Example: `/results/latest?format=json`.
  * **GET `/results/date/{date}`**: Searches for a result on a specific date. The date format is `YYYY-MM-DD`. Example: `/results/date/2024-01-15`.
  * **GET `/results/year/{year}`**: Returns all results for a specific year. The year format is `YYYY`. Example: `/results/year/2023`.
  * **GET `/results/month/{month}`**: Returns all results for a specific month and year. The month format is `YYYY-MM`. Example: `/results/month/2024-03`.

<hr> 


### Examples:

(Last)  
[https://api-euromillions.nunofcguerreiro.com/results/latest](https://api-euromillions.nunofcguerreiro.com/results/latest)  
[https://api-euromillions.nunofcguerreiro.com/results/latest?format=xml](https://api-euromillions.nunofcguerreiro.com/results/latest?format=xml)  
[https://api-euromillions.nunofcguerreiro.com/results/latest?format=plaintext](https://api-euromillions.nunofcguerreiro.com/results/latest?format=plaintext)  

(All)  
[https://api-euromillions.nunofcguerreiro.com/results](https://api-euromillions.nunofcguerreiro.com/results)  
[https://api-euromillions.nunofcguerreiro.com/results?format=xml](https://api-euromillions.nunofcguerreiro.com/results?format=xml)  
[https://api-euromillions.nunofcguerreiro.com/results?format=plaintext](https://api-euromillions.nunofcguerreiro.com/results?format=plaintext)  


(Date)  
[https://api-euromillions.nunofcguerreiro.com/results/date/2025-08-19](https://api-euromillions.nunofcguerreiro.com/results/date/2025-08-19)  
[https://api-euromillions.nunofcguerreiro.com/results/date/2025-08-19?format=xml](https://api-euromillions.nunofcguerreiro.com/results/date/2025-08-19?format=xml)  
[https://api-euromillions.nunofcguerreiro.com/results/date/2025-08-19?format=plaintext](https://api-euromillions.nunofcguerreiro.com/results/date/2025-08-19?format=plaintext)


(All Year)  
[https://api-euromillions.nunofcguerreiro.com/results/year/2025](https://api-euromillions.nunofcguerreiro.com/results/year/2025)  
[https://api-euromillions.nunofcguerreiro.com/results/year/2025?format=xml](https://api-euromillions.nunofcguerreiro.com/results/year/2025?format=xml)   
[https://api-euromillions.nunofcguerreiro.com/results/year/2025?format=plaintext](https://api-euromillions.nunofcguerreiro.com/results/year/2025?format=plaintext)


(All Year/Month)   
[https://api-euromillions.nunofcguerreiro.com/results/month/2025-02](https://api-euromillions.nunofcguerreiro.com/results/month/2025-02)  
[https://api-euromillions.nunofcguerreiro.com/results/month/2025-02?format=xml](https://api-euromillions.nunofcguerreiro.com/results/month/2025-02?format=xml)  
[https://api-euromillions.nunofcguerreiro.com/results/month/2025-02?format=plaintext](https://api-euromillions.nunofcguerreiro.com/results/month/2025-02?format=plaintext)  





